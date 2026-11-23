.PHONY: fmt vet lint test test-race cover bench fuzz integration tidy check

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

test-race:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

bench:
	go test -run=^$$ -bench=. -benchmem ./...

fuzz:
	go test -run=^$$ -fuzz=FuzzDecimalUnmarshal -fuzztime=30s ./...
	go test -run=^$$ -fuzz=FuzzCandleUnmarshal -fuzztime=30s ./...
	go test -run=^$$ -fuzz=FuzzIncludedUnmarshal -fuzztime=30s ./...

integration:
	GECKOTERMINAL_INTEGRATION=1 go test -tags=integration -v -timeout 10m ./...

tidy:
	go mod tidy

check: fmt vet lint test-race
