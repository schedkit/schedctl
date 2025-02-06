agent:
	go build -o schedctl main.go

.PHONY: format
format:
	go fmt ./...

all: agent

.PHONY: test lint
test:
	go test -v -p 1 -race ./...

lint:
	golangci-lint run

clean:
	rm -rf schedctl

