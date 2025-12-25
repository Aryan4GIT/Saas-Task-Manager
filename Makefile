.PHONY: help run build test clean docker-build docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make run          - Run the application locally"
	@echo "  make build        - Build the application binary"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

run:
	go run cmd/server/main.go

build:
	go build -o bin/saas-backend cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t saas-backend .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

db-migrate:
	psql -U postgres -d saas_db -f database/schema.sql

db-seed:
	psql -U postgres -d saas_db -f database/seed.sql
