.PHONY: build run up down logs migrate-up migrate-down migrate-version test

build:
	docker-compose build

up:
	docker-compose up --build 

down:
	docker-compose down -v

logs:
	docker-compose logs -f

migrate-up:
	docker-compose run --rm migrate

migrate-down:
	docker-compose run --rm migrate sh -c "/app/migrate -path /app/migrations -command down"

migrate-version:
	docker-compose run --rm migrate sh -c "/app/migrate -path /app/migrations -command version"

run-local:
	ENV_PATH=./configs/.env go run ./cmd/api/main.go

test:
	go test ./tests/pr 
	go test ./tests/team 
	go test ./tests/user 
