.PHONY: test checks build

default: test checks build

test:
	go test -v -cover ./...

build:
	GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm ./demo.go

checks:
	golangci-lint run
