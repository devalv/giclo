fmt:
	go mod tidy
	go fix -diff main.go
	golangci-lint run --fix main.go

build:
	$(MAKE) fmt
	go env -w CGO_ENABLED=0
	go env -w GOOS=linux
	go env -w GOARCH=amd64
	go build -o application main.go

build-deb:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make build-deb VERSION=1.2.3"; \
		exit 1; \
	fi
	PACKAGE_VERSION="$(VERSION)" ./devops/build-deb.sh

run:
	go run main.go

.PHONY: setup fmt test build cover run build-deb
