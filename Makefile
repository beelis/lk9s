.PHONY: build run test lint vet tidy clean

BINARY := bin/lk9s

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/lk9s

run:
	go run ./cmd/lk9s

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ dist/
