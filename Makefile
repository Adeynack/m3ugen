build:
	go build -o bin/m3ugen ./cmd/m3ugen/*.go

b: build

clean:
	go clean -cache -testcache

c: clean

install:
	go install ./cmd/m3ugen

i: install

test:
	go test ./...

t: test

ct: clean test

lint: build
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
