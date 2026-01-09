BINARY_NAME=huntsman

all: build

setup-hooks:
	git config core.hooksPath .githooks

build:
	go build -o $(BINARY_NAME) .

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 .

clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

run: build
	./$(BINARY_NAME)
	
install: build
ifeq ($(shell uname), Darwin)
	mkdir -p /usr/local/bin
	cp $(BINARY_NAME) /usr/local/bin/
	mkdir -p "$(HOME)/Library/Application Support/huntsman"
	cp theme.json "$(HOME)/Library/Application Support/huntsman/"
else
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/
	mkdir -p $(HOME)/.config/huntsman
	cp theme.json $(HOME)/.config/huntsman/
endif

.PHONY: all build clean run setup-hooks install
