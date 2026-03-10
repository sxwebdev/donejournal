package todos

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/pkg/broker"
)

// TodoEvent is published whenever a todo is created, updated, or deleted.
type TodoEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[TodoEvent]
}

// New creates a new todos service
func New(store *store.Store) *Service {
	b := broker.NewBroker[TodoEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the todos event broker.
func (s *Service) Broker() *broker.Broker[TodoEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
