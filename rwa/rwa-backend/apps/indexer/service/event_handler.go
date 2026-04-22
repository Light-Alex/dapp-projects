package service

import (
	"context"

	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"gorm.io/gorm"
)

type EventHandler interface {
	// ContractType returns the type of contract the handler processes.
	ContractType() coreTypes.ContractType

	// Topic0 returns the topic0 of the event the handler handles.
	Topic0() string

	// HandleEvent handles a single event log. The tx parameter is the database
	// transaction that the handler MUST use for all DB operations to ensure atomicity.
	HandleEvent(ctx context.Context, tx *gorm.DB, event *coreTypes.EventLogWithId) error
}

