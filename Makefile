SHELL = /bin/bash

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: integration-test
integration-test: deps
	go test -tags=integration ./...

.PHONY: coverage
coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: build
build:
	go build -o messagen.bin

.PHONY: cross-build-snapshot
cross-build:
	goreleaser --rm-dist --snapshot

.PHONY: install
install:
	go install

.PHONY: release
rellease:
	goreleaser release

.PHONY: build-image
build-image:
	docker build -t mpppk/messagen .

.PHONY: run-image-with-pokemon
run-image-with-pokemon:
	docker run -v ./examples/iroha/pokemon.yaml:/messagen.yaml -it mpppk/messagen