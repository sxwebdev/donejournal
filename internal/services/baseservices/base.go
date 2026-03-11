package baseservices

import (
	"github.com/sxwebdev/donejournal/internal/services/inbox"
	"github.com/sxwebdev/donejournal/internal/services/notes"
	"github.com/sxwebdev/donejournal/internal/services/todos"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/tkcrm/mx/logger"
)

type BaseServices struct {
	inboxService *inbox.Service
	todosService *todos.Service
	notesService *notes.Service
}

func New(
	l logger.Logger,
	st *store.Store,
) *BaseServices {
	inboxService := inbox.New(st)
	todosService := todos.New(st)
	notesService := notes.New(st)

	return &BaseServices{
		inboxService: inboxService,
		todosService: todosService,
		notesService: notesService,
	}
}

// Inbox returns inbox service
func (b *BaseServices) Inbox() *inbox.Service {
	return b.inboxService
}

// Todos returns todos service
func (b *BaseServices) Todos() *todos.Service {
	return b.todosService
}

// Notes returns notes service
func (b *BaseServices) Notes() *notes.Service {
	return b.notesService
}

// Stop stops all services and their brokers.
func (b *BaseServices) Stop() {
	b.todosService.Stop()
	b.inboxService.Stop()
	b.notesService.Stop()
}
