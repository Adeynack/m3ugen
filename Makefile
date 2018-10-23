build:
	GO111MODULE=on go build -o bin/m3ugen ./cmd/m3ugen/*.go

clean:
	GO111MODULE=on go clean -i -cache -testcache ./...

install:
	GO111MODULE=on go install ./cmd/m3ugen
