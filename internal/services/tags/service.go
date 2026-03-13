package tags

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/pkg/broker"
)

// TagEvent is published whenever a tag or tag association is created, updated, or deleted.
type TagEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[TagEvent]
}

// New creates a new tags service
func New(store *store.Store) *Service {
	b := broker.NewBroker[TagEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the tags event broker.
func (s *Service) Broker() *broker.Broker[TagEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
