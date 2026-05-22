format:
	go fmt ./...
	goimports -w .

lint:
	golangci-lint run

test:
	go test -race -coverprofile=./coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

run:
	go run cmd/miner/main.go
