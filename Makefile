MOCKERY := go run github.com/vektra/mockery/v2@v2.53.4

.PHONY: build test test-short e2e lint fmt cover verify mocks run clean

build: ## Build the binary to ./hookfire
	go build -o hookfire ./cmd/hookfire

test: ## Unit + integration tests with the race detector
	go test ./... -race -count=1

test-short: ## Fast tests, no race detector
	go test ./...

e2e: ## End-to-end tests against the built binary
	go test -tags e2e ./e2e/...

lint: ## go vet + golangci-lint (v2)
	go vet ./...
	golangci-lint run ./...

fmt: ## Format all Go files
	gofmt -w .

cover: ## Coverage summary (excludes generated mocks)
	go test ./internal/... -coverprofile=coverage.txt
	go tool cover -func=coverage.txt | tail -1

verify: build ## Confirm every built-in provider signs correctly
	./hookfire verify

mocks: ## Regenerate mocks (pinned mockery v2; PATH may shadow it with v3)
	$(MOCKERY)

run: build ## Build, then: make run ARGS="trigger github push --dry-run"
	./hookfire $(ARGS)

clean:
	rm -f hookfire coverage.txt
