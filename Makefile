.PHONY: build test lint clean install

build:
	go build -o bin/rgw ./cmd/rgw

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

install:
	go install ./cmd/rgw
