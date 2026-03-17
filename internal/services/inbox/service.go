package inbox

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/xutils/broker"
)

// InboxEvent is published whenever an inbox item is created, updated, or deleted.
type InboxEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[InboxEvent]
}

// New creates a new inbox service
func New(store *store.Store) *Service {
	b := broker.NewBroker[InboxEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the inbox event broker.
func (s *Service) Broker() *broker.Broker[InboxEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
