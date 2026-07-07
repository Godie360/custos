.PHONY: dev build test test-cover lint migrate-up migrate-down migrate-create generate tidy

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up

build:
	go build -o bin/server ./cmd/server

test:
	go test -race -count=1 ./...

test-cover:
	go test -race -coverprofile=cover.out ./... && go tool cover -html=cover.out

test-integration:
	go test -race -tags=integration ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

vuln:
	govulncheck ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

generate:
	go generate ./...

tidy:
	go mod tidy
