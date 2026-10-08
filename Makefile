BIN := devcheck

.PHONY: build test vet install

build:
	go build -o $(BIN) ./cmd/devcheck

test:
	go test ./...

vet:
	go vet ./...

install:
	go install ./cmd/devcheck
