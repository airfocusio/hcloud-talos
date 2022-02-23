.PHONY: *

test:
	go test -v ./...

test-watch:
	watch -n1 go test -v -timeout 60m ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --rm-dist --skip-publish --snapshot

release:
	goreleaser release --rm-dist
