.PHONY: all ebash test diff_test lint

all : ebash

ebash:
	@[ -f ebash ] || go build -o ebash ./cmd/ebash/main.go
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
