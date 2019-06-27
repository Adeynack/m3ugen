build:
	GO111MODULE=on go build -o bin/m3ugen ./cmd/m3ugen/*.go

b: build

clean:
	GO111MODULE=on go clean -i -cache -testcache ./...

c: clean

install:
	GO111MODULE=on go install ./cmd/m3ugen

i: install

test:
	GO111MODULE=on go test ./...

t: test

ct: clean test
