BINARY_NAME=huntsman

all: build

setup-hooks:
	git config core.hooksPath .githooks

build:
	go build -o $(BINARY_NAME) main.go

clean:
	go clean
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run setup-hooks
