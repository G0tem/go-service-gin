.PHONY: run build up down swagger

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

up:
	docker compose up -d --build

down:
	docker compose down -v

swagger:
	swag init -g cmd/server/main.go -o ./docs --parseInternal