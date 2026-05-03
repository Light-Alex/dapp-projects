package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/indexer/config"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

// EventListener listens to blockchain events directly via RPC
type EventListener struct {
	evmClient   *evm_helper.EvmClient
	conf        *config.Config
	chainID     uint64
	pocAddress  common.Address
	gateAddress common.Address
	// contractTypeByAddress maps contract addresses to their ContractType for event classification
	contractTypeByAddress map[common.Address]coreTypes.ContractType
}

func NewEventListener(evmClient *evm_helper.EvmClient, conf *config.Config) (*EventListener, error) {
	pocAddress := common.HexToAddress(conf.Chain.PocAddress)

	contractTypeByAddress := map[common.Address]coreTypes.ContractType{
		pocAddress: coreTypes.ContractTypeOrder,
	}

	var gateAddress common.Address
	if conf.Chain.GateAddress != "" {
		gateAddress = common.HexToAddress(conf.Chain.GateAddress)
		contractTypeByAddress[gateAddress] = coreTypes.ContractTypeGate
	}

	return &EventListener{
		evmClient:             evmClient,
		conf:                  conf,
		chainID:               conf.Chain.ChainId,
		pocAddress:            pocAddress,
		gateAddress:           gateAddress,
		contractTypeByAddress: contractTypeByAddress,
	}, nil
}

// monitoredAddresses returns the list of contract addresses to monitor.
func (l *EventListener) monitoredAddresses() []common.Address {
	addrs := []common.Address{l.pocAddress}
	if l.gateAddress != (common.Address{}) {
		addrs = append(addrs, l.gateAddress)
	}
	return addrs
}

// contractTypeForAddress returns the ContractType for a given address,
// defaulting to ContractTypeOrder if unknown.
func (l *EventListener) contractTypeForAddress(addr common.Address) coreTypes.ContractType {
	if ct, ok := l.contractTypeByAddress[addr]; ok {
		return ct
	}
	return coreTypes.ContractTypeOrder
}

// GetLatestBlockNumber returns the latest block number from the chain
func (l *EventListener) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	client := l.evmClient.MustGetHttpClient(l.chainID)
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.ErrorZ(ctx, "failed to get latest block number", zap.Error(err))
		return 0, err
	}
	return header.Number.Uint64(), nil
}

// FetchEventsByBlockRange fetches events for a range of blocks with proper event ID assignment
func (l *EventListener) FetchEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64, startEventID uint64) ([]*coreTypes.EventLogWithId, error) {
	client := l.evmClient.MustGetHttpClient(l.chainID)

	// Create filter query for all monitored contracts
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: l.monitoredAddresses(),
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		log.ErrorZ(ctx, "failed to filter logs",
			zap.String("rpcURL", l.evmClient.MustGetRpcInfo(l.chainID).RpcUrl),
			zap.Uint64("fromBlock", fromBlock),
			zap.Uint64("toBlock", toBlock),
			zap.Error(err))
		return nil, err
	}

	events := make([]*coreTypes.EventLogWithId, 0, len(logs))
	eventID := startEventID

	// Group logs by block to get block timestamps efficiently
	blockMap := make(map[uint64]*big.Int)
	for _, logEntry := range logs {
		if _, exists := blockMap[logEntry.BlockNumber]; !exists {
			header, err := client.HeaderByNumber(ctx, new(big.Int).SetUint64(logEntry.BlockNumber))
			if err != nil {
				log.ErrorZ(ctx, "failed to get block header",
					zap.Uint64("blockNumber", logEntry.BlockNumber),
					zap.Error(err))
				return nil, fmt.Errorf("failed to get block header for block %d: %w", logEntry.BlockNumber, err)
			}
			blockMap[logEntry.BlockNumber] = new(big.Int).SetUint64(header.Time)
		}
	}

	for _, logEntry := range logs {
		blockTime, ok := blockMap[logEntry.BlockNumber]
		if !ok {
			continue
		}

		// Convert topics to hex strings
		topics := make([]string, len(logEntry.Topics))
		for i, topic := range logEntry.Topics {
			topics[i] = topic.Hex()
		}

		eventID++
		event := &coreTypes.EventLogWithId{
			Address:        logEntry.Address,
			BlockHash:      logEntry.BlockHash.Hex(),
			BlockNumber:    hexutil.EncodeUint64(logEntry.BlockNumber),
			BlockTimestamp: hexutil.EncodeUint64(blockTime.Uint64()),
			Data:           hexutil.Encode(logEntry.Data),
			Index:          hexutil.EncodeUint64(uint64(logEntry.Index)),
			Topics:         topics,
			TxHash:         logEntry.TxHash.Hex(),
			TxIndex:        hexutil.EncodeUint64(uint64(logEntry.TxIndex)),
			Removed:        logEntry.Removed,
			EventId:        eventID,
			ContractType:   l.contractTypeForAddress(logEntry.Address),
		}
		events = append(events, event)
	}

	log.InfoZ(ctx, "fetched events from blockchain",
		zap.Uint64("fromBlock", fromBlock),
		zap.Uint64("toBlock", toBlock),
		zap.Uint64("startEventID", startEventID),
		zap.Int("eventCount", len(events)))

	return events, nil
}

// StartPolling starts polling for new blocks and events
func (l *EventListener) StartPolling(ctx context.Context, blockService *BlockService, processService *ProcessTxService) error {
	ticker := time.NewTicker(time.Duration(l.conf.Indexer.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.InfoZ(ctx, "stopping event listener polling")
			return nil
		case <-ticker.C:
			if err := l.pollAndProcess(ctx, blockService, processService); err != nil {
				log.ErrorZ(ctx, "error polling and processing events", zap.Error(err))
			}
		}
	}
}

func (l *EventListener) pollAndProcess(ctx context.Context, blockService *BlockService, processService *ProcessTxService) error {
	// Get latest block from chain
	latestBlock, err := l.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// Get last processed block
	lastBlock, err := blockService.GetLastProcessedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last processed block: %w", err)
	}

	// Apply confirmation blocks
	if l.conf.Indexer.ConfirmationBlocks > 0 {
		if latestBlock < l.conf.Indexer.ConfirmationBlocks {
			return nil
		}
		latestBlock = latestBlock - l.conf.Indexer.ConfirmationBlocks
	}

	if latestBlock <= lastBlock {
		return nil // No new blocks
	}

	// Process in batches
	batchSize := uint64(l.conf.Indexer.BatchSize)
	fromBlock := lastBlock + 1
	toBlock := lastBlock + batchSize
	if toBlock > latestBlock {
		toBlock = latestBlock
	}

	// Get last event ID
	lastEventID, err := blockService.GetLastEventID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last event ID: %w", err)
	}

	// Fetch events
	events, err := l.FetchEventsByBlockRange(ctx, fromBlock, toBlock, lastEventID)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		// No events, just update block number
		if err := blockService.UpdateLastProcessedBlockTx(ctx, blockService.db, toBlock, lastEventID); err != nil {
			return fmt.Errorf("failed to update last processed block: %w", err)
		}
		return nil
	}

	// Process all events and update block progress in a single atomic transaction
	if err := processService.ProcessBatch(ctx, events, toBlock, blockService); err != nil {
		return fmt.Errorf("failed to process event batch: %w", err)
	}

	log.InfoZ(ctx, "processed events",
		zap.Uint64("fromBlock", fromBlock),
		zap.Uint64("toBlock", toBlock),
		zap.Int("eventCount", len(events)),
		zap.Uint64("maxEventID", events[len(events)-1].EventId))

	return nil
}
