package repos

import (
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/store/repos/repo_inbox"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_notes"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_tags"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
)

type Repos struct {
	inbox      *repo_inbox.CustomQueries
	todos      *repo_todos.CustomQueries
	notes      *repo_notes.CustomQueries
	workspaces *repo_workspaces.CustomQueries
	tags       *repo_tags.CustomQueries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		inbox:      repo_inbox.NewCustom(sqlite),
		todos:      repo_todos.NewCustom(sqlite),
		notes:      repo_notes.NewCustom(sqlite),
		workspaces: repo_workspaces.NewCustom(sqlite),
		tags:       repo_tags.NewCustom(sqlite),
	}
}

// Inbox returns repo for requests
func (s *Repos) Inbox(opts ...Option) repo_inbox.ICustomQuerier {
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

// Notes returns repo for notes
func (s *Repos) Notes(opts ...Option) repo_notes.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.notes.WithTx(options.Tx)
	}

	return s.notes
}

// Workspaces returns repo for workspaces
func (s *Repos) Workspaces(opts ...Option) repo_workspaces.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.workspaces.WithTx(options.Tx)
	}

	return s.workspaces
}

// Tags returns repo for tags
func (s *Repos) Tags(opts ...Option) repo_tags.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.tags.WithTx(options.Tx)
	}

	return s.tags
}
