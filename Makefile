.PHONY: dev build install clean

dev:
	@echo "Starting development server..."
	@go run . & pnpm --dir frontend dev

build:
	@echo "Building frontend..."
	@pnpm --dir frontend build
	@echo "Building backend..."
	@go build -o docklog .

install: build
	@echo "Installing docklog to /usr/local/bin (may require sudo)..."
	@install -m 755 docklog /usr/local/bin/docklog
	@echo "Installed: docklog (run 'docklog help')"

docker-build:
	@echo "Building Docker image..."
	@docker build -t docklog:latest .

up:
	@echo "Starting DockLog and test containers..."
	@touch docklog.db
	@docker-compose up --build -d

down:
	@echo "Stopping containers..."
	@docker-compose down

clean:
	@rm -rf docklog frontend/dist
	@echo "Cleaned up."
