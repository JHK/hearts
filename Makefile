.PHONY: fmt test run-host run-cli

fmt:
	gofmt -w cmd internal

test:
	go test ./...

run-host:
	go run ./cmd/heartsd

run-cli:
	go run ./cmd/hearts-cli
