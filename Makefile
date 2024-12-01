# Simple Makefile for a Go project

# Build the application
all: build test

BUILD_DIR = vosskamp-reisen-3
BINARY_NAME = $(BUILD_DIR)

build:
	docker build --no-cache -t fedora-build .
	docker run --platform linux/amd64 -v $(shell pwd):/app fedora-build

build-docker: clean
	@echo ">>>>>>>>>>>> --------- Creating directory structure... --------- <<<<<<<<<<<<"
	@mkdir -p $(BUILD_DIR)/internal/database
	@mkdir -p $(BUILD_DIR)/internal/static
	@mkdir -p $(BUILD_DIR)/internal/templates

	@echo ">>>>>>>>>>>> --------- Building binary for Linux... --------- <<<<<<<<<<<<"
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/main ./cmd/api/main.go

	@echo ">>>>>>>>>>>> --------- Copying static files... --------- <<<<<<<<<<<<"
	@cp -r internal/static/* $(BUILD_DIR)/internal/static/ 2>/dev/null || true
	@cp -r internal/templates/* $(BUILD_DIR)/internal/templates/ 2>/dev/null || true
	@cp .env $(BUILD_DIR)/.env 2>/dev/null || true

	@echo ">>>>>>>>>>>> --------- Creating zip archive... --------- <<<<<<<<<<<<"
	@zip -r $(BUILD_DIR).zip $(BUILD_DIR)

	# @rm -rf $(BUILD_DIR)
	@echo ">>>>>>>>>>>> --------- Build complete! Output: $(BUILD_DIR).zip --------- <<<<<<<<<<<<"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BUILD_DIR).zip

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch
