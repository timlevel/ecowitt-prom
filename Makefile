APP_NAME := ecowitt-prom
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build run test clean docker

build:
	go build $(LDFLAGS) -o $(APP_NAME) .

run: build
	./$(APP_NAME)

test:
	go test -v ./...

clean:
	rm -f $(APP_NAME)

docker:
	docker build -t $(APP_NAME):$(VERSION) .