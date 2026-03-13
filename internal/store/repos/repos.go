package repos

import (
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/store/repos/repo_inbox"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_notes"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_projects"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
)

type Repos struct {
	inbox    *repo_inbox.CustomQueries
	todos    *repo_todos.CustomQueries
	notes    *repo_notes.CustomQueries
	projects *repo_projects.CustomQueries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		inbox:    repo_inbox.NewCustom(sqlite),
		todos:    repo_todos.NewCustom(sqlite),
		notes:    repo_notes.NewCustom(sqlite),
		projects: repo_projects.NewCustom(sqlite),
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

// Projects returns repo for projects
func (s *Repos) Projects(opts ...Option) repo_projects.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.projects.WithTx(options.Tx)
	}

	return s.projects
}
