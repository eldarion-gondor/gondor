guard-%:
	@ if [ "${${*}}" == "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi

build:
	godep go install -a -ldflags "-X main.version=0.0.0-dev" ./cmd/gondor

bin/gondor-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-linux-amd64 ./cmd/gondor

bin/gondor-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-darwin-amd64 ./cmd/gondor

bin/gondor-windows-amd64.exe:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-X main.version=$(VERSION)" -o bin/gondor-windows-amd64.exe ./cmd/gondor

cross-build: guard-VERSION bin/gondor-linux-amd64 bin/gondor-darwin-amd64 bin/gondor-windows-amd64.exe

release: guard-VERSION cross-build
	gsutil cp -a public-read ./bin/gondor-linux-amd64 gs://gondor-cli/gondor-v$(VERSION)-linux-amd64
	gsutil cp -a public-read ./bin/gondor-darwin-amd64 gs://gondor-cli/gondor-v$(VERSION)-darwin-amd64
	gsutil cp -a public-read ./bin/gondor-windows-amd64.exe gs://gondor-cli/gondor-v$(VERSION)-windows-amd64.exe

clean:
	rm -rf bin/

.PHONY: build cross-build release
