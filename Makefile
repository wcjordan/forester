.PHONY: build test lint run dev clean check format e2e_viz wasm road_preview

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

wasm:
	GOOS=js GOARCH=wasm go build -o forester.wasm .

clean:
	rm -f $(BINARY) forester.wasm
	rm -rf tmp/

check: lint test

format:
	gofmt -s -w .

TESTNAME := TestLogStorageWorkflow
e2e_viz:
	go clean -testcache
	E2E_VISUAL=1 E2E_VISUAL_DELAY=150ms go test ./e2e_tests/ -run $(TESTNAME) -v

road_preview:
	go run ./tools/road-preview