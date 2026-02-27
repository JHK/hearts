.PHONY: setup fmt test css css-watch run

setup:
	npm install
	go mod download

fmt:
	gofmt -w cmd internal

test:
	go test ./...

css:
	npm run build:css

css-watch:
	npm run watch:css

run: css
	go run ./cmd/hearts
