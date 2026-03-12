BINARY := gh-godo

.PHONY: all build test run clean

all: build

build:
	go build -o $(BINARY) .

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
