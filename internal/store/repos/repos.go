package repos

import (
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/store/repos/repo_inbox"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
)

type Repos struct {
	inbox *repo_inbox.Queries
	todos *repo_todos.CustomQueries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		inbox: repo_inbox.New(sqlite),
		todos: repo_todos.NewCustom(sqlite),
	}
}

// Inbox returns repo for requests
func (s *Repos) Inbox(opts ...Option) repo_inbox.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.inbox.WithTx(options.Tx)
	}

	return s.inbox
}

// Todos returns repo for todos
func (s *Repos) Todos(opts ...Option) repo_todos.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.todos.WithTx(options.Tx)
	}

	return s.todos
}
