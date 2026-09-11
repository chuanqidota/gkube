.PHONY: lint lint-fix build run test check

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

build:
	go build -o gkube .

run:
	go run main.go

test:
	go test ./...

check: lint build
