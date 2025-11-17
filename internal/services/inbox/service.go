package inbox

import "github.com/sxwebdev/donejournal/internal/store"

type Service struct {
	store *store.Store
}

// New creates a new inbox service
func New(store *store.Store) *Service {
	return &Service{
		store: store,
	}
}
