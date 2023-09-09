setup:
	echo "Install all the build and lint dependencies"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $(go env GOPATH)/bin v1.48.0
	go install golang.org/x/tools/cmd/goimports@latest

pcr:
	pre-commit run --all-files

lint:
	golangci-lint run -c .golangci.yml

fmt:
	gofmt -w -s ./internal
	goimports -w ./internal

test:
	go test ./... -race

cover:
	go test ./... -race -cover

build:
	$(MAKE) fmt
	$(MAKE) lint
	go build -o application ./cmd

run:
	go run ./cmd
