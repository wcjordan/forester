.PHONY: build test lint run dev clean check

BINARY := forester

build:
	go build -o $(BINARY) .

test:
	go test -race ./...

lint:
	golangci-lint run ./...

run: build
	./$(BINARY)

dev:
	air

clean:
	rm -f $(BINARY)
	rm -rf tmp/

check: lint test
