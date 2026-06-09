# Define the binary file names
BINARY_COORDINATOR=cdntClient
BINARY_TESTMANAGER=tmClient
BINARY_WORKER=workerClient
BINARY_EXPMANAGER=expInitAgent

# Define the source file paths
SRC_COORDINATOR=./clients/coordinator/main.go
SRC_TESTMANAGER=./clients/testManager/main.go
SRC_WORKER=./clients/worker/main.go
SRC_EXPMANAGER=./deployScripts/expManager/main.go

# Directory names
LOCALTEST_DIR=localtest
GENERALTEST_DIR=generaltest
EXPBULD_DIR=expOPL

# Default target: build all clients
.PHONY: all
all: build-coordinator build-testmanager build-worker

# Build coordinator
.PHONY: build-coordinator
build-coordinator:
	go build -o $(BINARY_COORDINATOR) $(SRC_COORDINATOR)

# Build testmanager
.PHONY: build-testmanager
build-testmanager:
	go build -o $(BINARY_TESTMANAGER) $(SRC_TESTMANAGER)

# Build worker
.PHONY: build-worker
build-worker:
	go build -o $(BINARY_WORKER) $(SRC_WORKER)

# Build expInitAgent (expManager)
.PHONY: build-expmanager
build-expmanager:
	go build -o $(EXPBULD_DIR)/$(BINARY_EXPMANAGER) $(SRC_EXPMANAGER)

# Clean generated binary files
.PHONY: clean
clean:
	rm -rf $(LOCALTEST_DIR) $(GENERALTEST_DIR) $(EXPBULD_DIR)
	rm -f $(BINARY_COORDINATOR) $(BINARY_TESTMANAGER) $(BINARY_WORKER) $(BINARY_EXPMANAGER)

# localtest target: create or clean the localtest directory and build clients
.PHONY: localtest
localtest:
	@if [ -d $(LOCALTEST_DIR) ]; then \
		rm -rf $(LOCALTEST_DIR); \
		echo "Removed existing $(LOCALTEST_DIR) directory"; \
	else \
		echo "No existing $(LOCALTEST_DIR) directory found"; \
	fi
	@mkdir $(LOCALTEST_DIR)
	@echo "Created $(LOCALTEST_DIR) directory"
	go build -o $(LOCALTEST_DIR)/$(BINARY_COORDINATOR) $(SRC_COORDINATOR)
	go build -o $(LOCALTEST_DIR)/$(BINARY_TESTMANAGER) $(SRC_TESTMANAGER)
	go build -o $(LOCALTEST_DIR)/$(BINARY_WORKER) $(SRC_WORKER)
	@echo "Built $(BINARY_COORDINATOR), $(BINARY_TESTMANAGER), and $(BINARY_WORKER) in $(LOCALTEST_DIR) directory"
	@go run deployScripts/localtest/main.go $(LOCALTEST_DIR)
	@echo "Ran deploy script for $(LOCALTEST_DIR)"

# generaltest target: create or clean the generaltest directory and build clients
.PHONY: generaltest
generaltest:
	@if [ -d $(GENERALTEST_DIR) ]; then \
		rm -rf $(GENERALTEST_DIR); \
		echo "Removed existing $(GENERALTEST_DIR) directory"; \
	else \
		echo "No existing $(GENERALTEST_DIR) directory found"; \
	fi
	@mkdir $(GENERALTEST_DIR)
	@echo "Created $(GENERALTEST_DIR) directory"
	go build -o $(GENERALTEST_DIR)/$(BINARY_COORDINATOR) $(SRC_COORDINATOR)
	go build -o $(GENERALTEST_DIR)/$(BINARY_TESTMANAGER) $(SRC_TESTMANAGER)
	go build -o $(GENERALTEST_DIR)/$(BINARY_WORKER) $(SRC_WORKER)
	@echo "Built $(BINARY_COORDINATOR), $(BINARY_TESTMANAGER), and $(BINARY_WORKER) in $(GENERALTEST_DIR) directory"
	@go run deployScripts/generaltest/main.go $(GENERALTEST_DIR)
	@echo "Ran deploy script for $(GENERALTEST_DIR)"

# expbuild target: create expOPL directory, build clients and expInitAgent, and copy files
.PHONY: expbuild
expbuild:
	@if [ -d $(EXPBULD_DIR) ]; then \
		rm -rf $(EXPBULD_DIR); \
		echo "Removed existing $(EXPBULD_DIR) directory"; \
	else \
		echo "No existing $(EXPBULD_DIR) directory found"; \
	fi
	@mkdir $(EXPBULD_DIR)
	@echo "Created $(EXPBULD_DIR) directory"
	go build -o $(EXPBULD_DIR)/$(BINARY_COORDINATOR) $(SRC_COORDINATOR)
	go build -o $(EXPBULD_DIR)/$(BINARY_TESTMANAGER) $(SRC_TESTMANAGER)
	go build -o $(EXPBULD_DIR)/$(BINARY_WORKER) $(SRC_WORKER)
	go build -o $(EXPBULD_DIR)/$(BINARY_EXPMANAGER) $(SRC_EXPMANAGER)
	@echo "Built $(BINARY_COORDINATOR), $(BINARY_TESTMANAGER), $(BINARY_WORKER), and $(BINARY_EXPMANAGER) in $(EXPBULD_DIR) directory"
	@cp ./deployScripts/expManager/initExp.yaml $(EXPBULD_DIR)/
	@cp ./deployScripts/expManager/README.md $(EXPBULD_DIR)/
	@echo "Copied initExp.yaml and README.md to $(EXPBULD_DIR)"

# Help information
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo "Targets:"
	@echo "  all               - Build all clients (default target)"
	@echo "  build-coordinator - Build coordinator"
	@echo "  build-testmanager - Build testmanager"
	@echo "  build-worker      - Build worker"
	@echo "  clean             - Clean generated binary files"
	@echo "  localtest         - Create or clean localtest directory and build clients"
	@echo "  generaltest       - Create or clean generaltest directory and build clients"
	@echo "  expbuild          - Create expOPL directory, build clients and expInitAgent, and copy files"
	@echo "  help              - Display this help information"
