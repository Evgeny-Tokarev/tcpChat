SERVER_TCP_PORT=13003

PHONY: run-server
run-server:
	export SERVER_PORT=$(SERVER_TCP_PORT) && go run ./cmd/server/