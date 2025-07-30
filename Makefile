APP_NAME := pharmacy-claims-app

.PHONY: help build run test clean setup stop shell db-shell

help:
	@echo "Pharmacy Claims Application - Makefile"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make run         - Start development environment with live reload"
	@echo ""
	@echo "📋 Available targets:"
	@echo "  run          - Start development environment with live reload"
	@echo "  test         - Run unit tests"
	@echo "  clean        - Clean up containers and volumes"
	@echo "  stop         - Stop all services"
	@echo "  shell        - Open a development shell with Go tools"
	@echo "  db-shell     - Connect to PostgreSQL shell"

run:
	@echo "Starting $(APP_NAME)..."
	@echo "This will mount your source code for live development"
	@LOG_LEVEL=debug GO_ENV=development docker-compose up --build
	@echo "Application started"
	@echo "API available at: http://localhost:8080"
	@echo "Code changes will trigger automatic rebuilds"

stop:
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped"

test:
	@echo "Running unit tests..."
	@docker build -t $(APP_NAME):test .
	@docker run --rm -v $(PWD):/app -w /app $(APP_NAME):test go test -v ./...

clean:
	@echo "Cleaning up..."
	@docker-compose down -v --remove-orphans
	@docker rmi $(APP_NAME):latest $(APP_NAME):test 2>/dev/null || true
	@rm -f main
	@echo "Cleanup complete"

shell:
	@echo "Opening development shell..."
	@docker run --rm -it -v $(PWD):/app -w /app $(APP_NAME):latest sh

db-shell:
	@echo "Connecting to PostgreSQL..."
	@docker-compose exec postgres psql -U pharmacy_user -d pharmacy_claims
