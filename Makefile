.PHONY: *

test:
	go test -v $(go list ./... | grep -v e2etests)

test-e2e:
	go test -v $(go list ./... | grep e2etests) -timeout 60m

build:
	goreleaser release --rm-dist --skip-publish --snapshot

release:
	goreleaser release --rm-dist
