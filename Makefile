.PHONY: run lint lint-check-deps


GOPKGS = $(shell go list ./... | grep -v /vendor/)

run: 
	@echo "[ingester] running with default params"
	@go run main.go

build:
	go build -o ingester

test:
	go test -v $(GOPKGS)