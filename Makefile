run:
	go run ./cmd/server

test:
	go test ./...

migrate-up:
	go run ./cmd/migrate --direction=up

migrate-down:
	go run ./cmd/migrate --direction=down
