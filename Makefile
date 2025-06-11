# Variables
DOCKER_COMPOSE = docker compose
PROJECT_NAME = auctionEngine

# Comandos principales
.PHONY: run
run:
	@echo "Starting all services with Docker Compose..."
	$(DOCKER_COMPOSE) up --build -d

.PHONY: stop
stop:
	@echo "Stopping all services..."
	$(DOCKER_COMPOSE) down

.PHONY: restart
restart:
	@echo "Restarting all services..."
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up --build -d

.PHONY: logs
logs:
	@echo "Showing logs for all services..."
	$(DOCKER_COMPOSE) logs -f

.PHONY: clean
clean:
	@echo "Stopping and removing all containers, networks, and volumes..."
	$(DOCKER_COMPOSE) down -v

.PHONY: db-shell
db-shell:
	@echo "Opening a shell to the PostgreSQL database container..."
	docker exec -it auctionengine-db-1 psql -U $(shell grep DB_USER .env | cut -d '=' -f2) -d $(shell grep DB_NAME .env | cut -d '=' -f2)

.PHONY: seed-sql
seed-sql:
	@echo "Running SQL seeder script..."
	docker exec -i auctionengine-db-1 psql -U $(shell grep DB_USER .env | cut -d '=' -f2) -d $(shell grep DB_NAME .env | cut -d '=' -f2) < seed_data.sql

.PHONY: api-shell
api-shell:
	@echo "Opening a shell to the REST API container..."
	docker exec -it rest-books-server sh

.PHONY: build
build:
	@echo "Building the Docker images..."
	$(DOCKER_COMPOSE) build

.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run


.PHONY: help
help:
	@echo "Available commands:"
	@echo "  run         - Start all services with Docker Compose"
	@echo "  stop        - Stop all services"
	@echo "  restart     - Restart all services"
	@echo "  logs        - Show logs for all services"
	@echo "  clean       - Stop and remove all containers, networks, and volumes"
	@echo "  db-shell    - Open a shell to the PostgreSQL database container"
	@echo "  api-shell   - Open a shell to the REST API container"
	@echo "  build       - Build the Docker images"
	@echo "  test        - Run Go tests"
	@echo "  lint        - Run the linter"
	@echo "  migrate     - Run database migrations"