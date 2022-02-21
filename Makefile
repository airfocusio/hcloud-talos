.PHONY: *

run-bootstrap-cluster:
	mkdir -p test
	cd test && go run .. bootstrap-cluster --cluster-name=cluster --node-name=controlplane-01 --force

run-destroy-cluster:
	mkdir -p test
	cd test && go run .. destroy-cluster --force

run-apply-manifests:
	cd test && go run .. apply-manifests

run-add-node:
	cd test && go run .. add-node --node-name=worker-01

test:
	go test -v ./...

test-watch:
	watch -n1 go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --rm-dist --skip-publish --snapshot

release:
	goreleaser release --rm-dist
