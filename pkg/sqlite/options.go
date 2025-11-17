package sqlite

type Option func(*SQLite)

func WithName(name string) Option {
	return func(sq *SQLite) {
		sq.name = name
	}
}
