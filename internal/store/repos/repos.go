package repos

import (
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/store/repos/repo_requests"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
)

type Repos struct {
	requests *repo_requests.Queries
	todos    *repo_todos.Queries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		requests: repo_requests.New(sqlite),
		todos:    repo_todos.New(sqlite),
	}
}

// Requests returns repo for requests
func (s *Repos) Requests(opts ...Option) repo_requests.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.requests.WithTx(options.Tx)
	}

	return s.requests
}

// Todos returns repo for todos
func (s *Repos) Todos(opts ...Option) repo_todos.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.todos.WithTx(options.Tx)
	}

	return s.todos
}
