BINARY_NAME=huntsman

all: build

build:
	go build -o $(BINARY_NAME) main.go

clean:
	go clean
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run
