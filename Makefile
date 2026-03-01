.PHONY: build test lint run dev clean check format e2e_viz

BINARY := forester

build:
	go build -o $(BINARY) .

test:
	go test -race ./...

lint:
	golangci-lint run ./...
	go vet ./...

run: build
	./$(BINARY)

dev:
	air

clean:
	rm -f $(BINARY)
	rm -rf tmp/

check: lint test

format:
	gofmt -s -w .

TESTNAME := TestLogStorageWorkflow
e2e_viz:
	go clean -testcache
	E2E_VISUAL=1 E2E_VISUAL_DELAY=150ms go test ./e2e_tests/ -run $(TESTNAME) -v
