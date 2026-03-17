package workspaces

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/xutils/broker"
)

// WorkspaceEvent is published whenever a workspace is created, updated, or deleted.
type WorkspaceEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[WorkspaceEvent]
}

// New creates a new workspaces service
func New(store *store.Store) *Service {
	b := broker.NewBroker[WorkspaceEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the workspaces event broker.
func (s *Service) Broker() *broker.Broker[WorkspaceEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
