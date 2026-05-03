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

.PHONY: man man-check
man:
	go run -tags containers_image_openpgp ./cmd/gen-man -out dist/man

man-check:
	@tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	go run -tags containers_image_openpgp ./cmd/gen-man -out "$$tmp" || exit $$?; \
	diff -ruN dist/man "$$tmp" || { \
		echo; \
		echo 'manpages out of date — run `make man` and commit the result'; \
		exit 1; \
	}

clean:
	rm -rf schedctl

