GOPATH=$(shell pwd)/Godeps/_workspace
VERSION ?= $(shell cat VERSION | tr -d '\n ')

build:
	GOPATH=$(GOPATH) go build -a -ldflags "-X main.version=$(VERSION)" -o bin/g3a ./src

bin/g3a-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(GOPATH) go build -ldflags "-X main.version=$(VERSION)" -o bin/g3a-linux-amd64 ./src

bin/g3a-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(GOPATH) go build -ldflags "-X main.version=$(VERSION)" -o bin/g3a-darwin-amd64 ./src

bin/g3a-windows-amd64.exe:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(GOPATH) go build -ldflags "-X main.version=$(VERSION)" -o bin/g3a-windows-amd64.exe ./src

cross-build: bin/g3a-linux-amd64 bin/g3a-darwin-amd64 bin/g3a-windows-amd64.exe

release: cross-build
	gsutil cp -a public-read ./bin/g3a-linux-amd64 gs://gondor-cli/g3a-v$(VERSION)-linux-amd64
	gsutil cp -a public-read ./bin/g3a-darwin-amd64 gs://gondor-cli/g3a-v$(VERSION)-darwin-amd64
	gsutil cp -a public-read ./bin/g3a-windows-amd64.exe gs://gondor-cli/g3a-v$(VERSION)-windows-amd64.exe

clean:
	rm -rf bin/

.PHONY: build cross-build release
