.PHONY: all ebash test lint

all : ebash

ebash:
	@[ -f ebash ] || go build -o ebash ./cmd/ebash/main.go
	@./ebash

clean:
	@rm -f ./ebash
	@rm -f .ebash_history

test:
	go test -v ./...

lint:
	golangci-lint run ./...
