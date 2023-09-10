setup:
	pip install pre-commit
	pre-commit install

pcr:
	pre-commit autoupdate
	pre-commit run --all-files

fmt:
	gofmt -w -s ./internal
	goimports -w ./internal

test:
	go test ./... -race

cover:
	go test ./... -race -cover

build:
	$(MAKE) fmt
	go build -o application ./cmd

run:
	go run ./cmd
