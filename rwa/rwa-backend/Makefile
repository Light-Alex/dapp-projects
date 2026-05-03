# The binary to build (just the basename).
BIN := rwa-backend
PROJECT_NAME := "rwa-backend"

all: lint sync

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yaml -v ./apps/api
	golangci-lint run -c .golangci.yaml -v ./libs/database
	golangci-lint run -c .golangci.yaml -v ./libs/grpc/tx_service
	golangci-lint run -c .golangci.yaml -v ./libs/log
	golangci-lint run -c .golangci.yaml -v ./libs/errors
	golangci-lint run -c .golangci.yaml -v ./libs/kafka

.PHONY: sync
sync:
	go work sync

# Declare phony targets for installers and install_all
.PHONY: install_kafka install_database install_all

install_kafka:
	docker compose -f ./devops/local/kafka/docker-compose.yml up -d

install_database:
	docker compose -f ./devops/local/database/docker-compose.yml up -d

# install_all will run both installers
install_all: install_kafka install_database
