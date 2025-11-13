air:
	air -c .air.toml

start:
	go run ./cmd/donejournal start -c config.yaml

genenvs:
	go run ./cmd/donejournal config genenvs
