BINARY_NAME=donejournal
APP_PATH=./cmd/donejournal

air:
	air -c .air.toml

start:
	go run ./cmd/donejournal start -c config.yaml

# Migrations

migrateup:
	go run $(APP_PATH) migrations up -db-path ./data/sqlite/db.sqlite

migratedown:
	go run $(APP_PATH) migrations down -db-path ./data/sqlite/db.sqlite

migratecreate:
	go run $(APP_PATH) migrations create -p ./sql/migrations -name $(filter-out $@,$(MAKECMDGOALS))

# Generate

genenvs:
	go run ./cmd/donejournal config genenvs

gensql:
	pgxgen --pgxgen-config sql/pgxgen.yaml --sqlc-config sql/sqlc.yaml crud
	pgxgen --pgxgen-config sql/pgxgen.yaml --sqlc-config sql/sqlc.yaml sqlc generate
