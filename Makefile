
GIT_COMMIT = $(shell git rev-parse HEAD | cut -c 1-8)

build:
	godep go install -a -ldflags "-X main.buildSHA=dev" ./cmd/gondor

binaries:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.buildSHA=$(GIT_COMMIT)" -o bin/gondor-linux-amd64-$(GIT_COMMIT) ./cmd/gondor
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.buildSHA=$(GIT_COMMIT)" -o bin/gondor-darwin-amd64-$(GIT_COMMIT) ./cmd/gondor
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.buildSHA=$(GIT_COMMIT)" -o bin/gondor-windows-amd64-$(GIT_COMMIT).exe ./cmd/gondor
