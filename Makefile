GOPATH=$(shell pwd)/Godeps/_workspace
VERSION=$(shell cat VERSION | tr -d '\n ')

build:
	go build -a -ldflags "-X main.version=$(VERSION)" -o bin/gondor ./src

bin/gondor-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-linux-amd64 ./src

bin/gondor-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-darwin-amd64 ./src

bin/gondor-windows-amd64.exe:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-windows-amd64.exe ./src

cross-build: bin/gondor-linux-amd64 bin/gondor-darwin-amd64 bin/gondor-windows-amd64.exe

release: cross-build
	gsutil cp -a public-read ./bin/gondor-linux-amd64 gs://gondor-cli/gondor-v$(VERSION)-linux-amd64
	gsutil cp -a public-read ./bin/gondor-darwin-amd64 gs://gondor-cli/gondor-v$(VERSION)-darwin-amd64
	gsutil cp -a public-read ./bin/gondor-windows-amd64.exe gs://gondor-cli/gondor-v$(VERSION)-windows-amd64.exe

clean:
	rm -rf bin/

.PHONY: build cross-build release
