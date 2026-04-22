# The binary to build (just the basename).
BIN := rwa-api
PROJECT_NAME := "rwa-api"

all: swagger

.PHONY: swagger
swagger:
	swag i -g server/router.go --parseDependency --parseInternal -o docs

build:
	go build -o $(BIN) main.go