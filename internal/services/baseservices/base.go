package baseservices

import (
	"github.com/sxwebdev/donejournal/internal/services/inbox"
	"github.com/sxwebdev/donejournal/internal/services/todos"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/tkcrm/mx/logger"
)

type BaseServices struct {
	inboxService *inbox.Service
	todosService *todos.Service
}

func New(
	l logger.Logger,
	st *store.Store,
) *BaseServices {
	inboxService := inbox.New(st)
	todosService := todos.New(st)

	return &BaseServices{
		inboxService: inboxService,
		todosService: todosService,
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
