package projects

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/pkg/broker"
)

// ProjectEvent is published whenever a project is created, updated, or deleted.
type ProjectEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[ProjectEvent]
}

// New creates a new projects service
func New(store *store.Store) *Service {
	b := broker.NewBroker[ProjectEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the projects event broker.
func (s *Service) Broker() *broker.Broker[ProjectEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
