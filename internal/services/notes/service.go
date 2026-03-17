package notes

import (
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/xutils/broker"
)

// NoteEvent is published whenever a note is created, updated, or deleted.
type NoteEvent struct {
	UserID int64
}

type Service struct {
	store  *store.Store
	broker *broker.Broker[NoteEvent]
}

// New creates a new notes service
func New(store *store.Store) *Service {
	b := broker.NewBroker[NoteEvent]()
	go b.Start()
	return &Service{
		store:  store,
		broker: b,
	}
}

// Broker returns the notes event broker.
func (s *Service) Broker() *broker.Broker[NoteEvent] {
	return s.broker
}

// Stop stops the broker.
func (s *Service) Stop() {
	s.broker.Stop()
}
