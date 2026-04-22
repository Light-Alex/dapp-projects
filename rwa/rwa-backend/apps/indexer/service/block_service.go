package service

import (
	"context"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/indexer/config"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BlockService maintains the latest processed block number
type BlockService struct {
	db    *gorm.DB
	conf  *config.Config
	chainID uint64
}

func NewBlockService(db *gorm.DB, conf *config.Config) *BlockService {
	return &BlockService{
		db:      db,
		conf:    conf,
		chainID: conf.Chain.ChainId,
	}
}

// GetLastProcessedBlock returns the last processed block number
func (s *BlockService) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	q, u := gplus.NewQuery[rwa.EventClientRecord]()
	q.Eq(&u.ChainID, s.chainID)
	recordList, dbRes := gplus.SelectList(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to select event_client_record", zap.Uint64("chainId", s.chainID), zap.Error(dbRes.Error))
		return 0, dbRes.Error
	}
	if recordList != nil && len(recordList) > 0 {
		return recordList[0].LastBlock, nil
	}
	// Initialize if not exists
	startBlock := s.conf.Indexer.StartBlock
	dbRes = gplus.Insert(&rwa.EventClientRecord{
		ChainID:     s.chainID,
		LastBlock:   startBlock,
		LastEventID: 0,
		UpdateAt:    time.Now(),
	}, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to insert event_client_record", zap.Uint64("chainId", s.chainID), zap.Error(dbRes.Error))
		return 0, dbRes.Error
	}
	return startBlock, nil
}

// UpdateLastProcessedBlockTx updates the last processed block number using the provided db or transaction.
func (s *BlockService) UpdateLastProcessedBlockTx(ctx context.Context, tx *gorm.DB, blockNum uint64, eventID uint64) error {
	q, u := gplus.NewQuery[rwa.EventClientRecord]()
	q.Eq(&u.ChainID, s.chainID).
		Set(&u.LastBlock, blockNum).
		Set(&u.LastEventID, eventID).
		Set(&u.UpdateAt, time.Now())
	dbRes := gplus.Update(q, gplus.Db(tx))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to update event_client_record in tx",
			zap.Uint64("chainId", s.chainID),
			zap.Uint64("blockNum", blockNum),
			zap.Uint64("eventID", eventID),
			zap.Error(dbRes.Error))
		return dbRes.Error
	}
	return nil
}

// GetLastEventID returns the last processed event ID
func (s *BlockService) GetLastEventID(ctx context.Context) (uint64, error) {
	q, u := gplus.NewQuery[rwa.EventClientRecord]()
	q.Eq(&u.ChainID, s.chainID)
	recordList, dbRes := gplus.SelectList(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to select event_client_record", zap.Uint64("chainId", s.chainID), zap.Error(dbRes.Error))
		return 0, dbRes.Error
	}
	if recordList != nil && len(recordList) > 0 {
		return recordList[0].LastEventID, nil
	}
	return 0, nil
}

