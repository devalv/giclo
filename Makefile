setup:
	go install golang.org/x/tools/gopls@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/orijtech/structslop/cmd/structslop@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/sqs/goreturns@latest
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
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
	go build -o app ./cmd

run:
	go run ./cmd
