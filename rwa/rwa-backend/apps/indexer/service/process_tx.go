package service

import (
	"context"

	"github.com/AnchoredLabs/rwa-backend/apps/indexer/config"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProcessTxService processes events with idempotency checks
type ProcessTxService struct {
	db                    *gorm.DB
	handlerMap            map[types.ContractType]map[string]EventHandler
	conf                  *config.Config
	latestResolvedEventId uint64
	initialized           bool
}

func NewProcessTxService(handlers []EventHandler, db *gorm.DB, conf *config.Config) *ProcessTxService {
	handlerMap := make(map[types.ContractType]map[string]EventHandler)
	for _, h := range handlers {
		if !lo.HasKey(handlerMap, h.ContractType()) {
			handlerMap[h.ContractType()] = make(map[string]EventHandler)
		}
		handlerMap[h.ContractType()][h.Topic0()] = h
	}
	return &ProcessTxService{
		db:                    db,
		handlerMap:            handlerMap,
		conf:                  conf,
		latestResolvedEventId: 0,
		initialized:           false,
	}
}

// InitCache initializes the service and loads the last processed event ID
func (s *ProcessTxService) InitCache(ctx context.Context) error {
	if s.initialized {
		return nil
	}
	// Check event_client_record exists
	q, u := gplus.NewQuery[rwa.EventClientRecord]()
	q.Eq(&u.ChainID, s.conf.Chain.ChainId)
	recordList, dbRes := gplus.SelectList(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to select event_client_record", zap.Uint64("chainId", s.conf.Chain.ChainId), zap.Error(dbRes.Error))
		return dbRes.Error
	}
	if recordList != nil && len(recordList) > 0 {
		s.latestResolvedEventId = recordList[0].LastEventID
	} else {
		s.latestResolvedEventId = 0
	}
	s.initialized = true
	return nil
}

// ProcessEvent processes a single event with idempotency check using the provided transaction.
// The resolvedUpTo parameter tracks the highest event ID resolved so far (including within the current batch),
// so that duplicate events within the same batch are also skipped. The caller must not rely on
// s.latestResolvedEventId being updated here; it is only updated after the transaction commits.
func (s *ProcessTxService) ProcessEvent(ctx context.Context, tx *gorm.DB, event *types.EventLogWithId, resolvedUpTo uint64) error {
	// Idempotency check: skip if event ID is already processed
	if event.EventId <= resolvedUpTo {
		log.InfoZ(ctx, "event already processed, skipping",
			zap.Uint64("eventId", event.EventId),
			zap.Uint64("resolvedUpTo", resolvedUpTo))
		return nil
	}

	// Check if handler exists for this event
	if !lo.HasKey(s.handlerMap, event.ContractType) || !lo.HasKey(s.handlerMap[event.ContractType], event.Topics[0]) {
		log.InfoZ(ctx, "no handler for event",
			zap.String("txHash", event.TxHash),
			zap.String("blockNumber", event.BlockNumber),
			zap.Uint64("eventId", event.EventId),
			zap.String("contractType", string(event.ContractType)),
			zap.String("topic0", event.Topics[0]))
		return nil
	}

	handler := s.handlerMap[event.ContractType][event.Topics[0]]
	err := handler.HandleEvent(ctx, tx, event)
	if err != nil {
		log.ErrorZ(ctx, "failed to handle event",
			zap.String("txHash", event.TxHash),
			zap.String("blockNumber", event.BlockNumber),
			zap.Uint64("eventId", event.EventId),
			zap.Error(err))
		return err
	}

	log.InfoZ(ctx, "event processed successfully",
		zap.Uint64("eventId", event.EventId),
		zap.String("txHash", event.TxHash),
		zap.String("blockNumber", event.BlockNumber))

	return nil
}

// ProcessBatch processes all events and updates block progress in a single atomic transaction.
// s.latestResolvedEventId is only updated after the transaction commits successfully.
func (s *ProcessTxService) ProcessBatch(ctx context.Context, events []*types.EventLogWithId, toBlock uint64, blockService *BlockService) error {
	maxEventID := s.latestResolvedEventId
	for _, event := range events {
		if event.EventId > maxEventID {
			maxEventID = event.EventId
		}
	}

	// Track the resolved event ID within this batch without mutating s.latestResolvedEventId.
	batchResolvedUpTo := s.latestResolvedEventId

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := s.ProcessEvent(ctx, tx, event, batchResolvedUpTo); err != nil {
				return err
			}
			if event.EventId > batchResolvedUpTo {
				batchResolvedUpTo = event.EventId
			}
		}
		// Update block progress in the same transaction
		return blockService.UpdateLastProcessedBlockTx(ctx, tx, toBlock, maxEventID)
	})

	if err != nil {
		return err
	}

	// Transaction committed successfully — now update the in-memory state.
	s.latestResolvedEventId = maxEventID

	return nil
}

