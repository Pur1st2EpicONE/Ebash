.PHONY: all ebash clean test lint

all : ebash

ebash: ./cmd/ebash/main.go $(shell find internal -name "*.go")
	@go build -o ebash ./cmd/ebash/main.go
	@./ebash

clean:
	@rm -f ./ebash
	@rm -f .ebash_history

test:
	@[ -f ebash ] || go build -o ebash ./cmd/ebash/main.go
	@bash diff_tests.sh
	@rm -f ./ebash
	@rm -f .ebash_history

lint:
	golangci-lint run ./...
