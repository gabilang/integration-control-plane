BINARY_NAME=platform-api
CMD_PATH=./cmd/platform-api

.PHONY: build run clean tidy

build:
	go build -o bin/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

clean:
	rm -rf bin/

tidy:
	go mod tidy
