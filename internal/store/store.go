package store

import (
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/store/badgerdb"
	"github.com/sxwebdev/donejournal/internal/store/repos"
)

type Store struct {
	*repos.Repos

	cache *badgerdb.DB

	sqlite *sql.DB
}

func New(sqlite *sql.DB, badgerDB *badgerdb.DB) (*Store, error) {
	return &Store{
		Repos:  repos.New(sqlite),
		sqlite: sqlite,
		cache:  badgerDB,
	}, nil
}

func (s *Store) SQLite() *sql.DB { return s.sqlite }

func (s *Store) Cache() *badgerdb.DB { return s.cache }
