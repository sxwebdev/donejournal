package baseservices

import (
	"github.com/sxwebdev/donejournal/internal/services/requests"
	"github.com/sxwebdev/donejournal/internal/services/todos"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/tkcrm/mx/logger"
)

type BaseServices struct {
	requestsService *requests.Service
	todosService    *todos.Service
}

func New(
	l logger.Logger,
	st *store.Store,
) *BaseServices {
	requestsService := requests.New(st)
	todosService := todos.New(st)

	return &BaseServices{
		requestsService: requestsService,
		todosService:    todosService,
	}
}

// Requests returns requests service
func (b *BaseServices) Requests() *requests.Service {
	return b.requestsService
}

// Todos returns todos service
func (b *BaseServices) Todos() *todos.Service {
	return b.todosService
}
