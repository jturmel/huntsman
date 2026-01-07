BINARY_NAME=huntsman

all: build

setup-hooks:
	git config core.hooksPath .githooks

build:
	go build -o $(BINARY_NAME) main.go

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 main.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 main.go

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 main.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 main.go

clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run setup-hooks
