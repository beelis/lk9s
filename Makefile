.PHONY: build run test e2e lint vet tidy clean

BINARY := bin/lk9s

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/lk9s

run:
	go run ./cmd/lk9s

test:
	go test -race -cover ./...

# Runs against an in-process fake LiveKit API. To use a real server, start one
# (docker run --rm --network host livekit/livekit-server --dev --bind 127.0.0.1)
# and run with LK9S_E2E_URL=http://127.0.0.1:7880.
e2e:
	go test -race -tags e2e -count=1 ./internal/lk/...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ dist/
