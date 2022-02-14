.PHONY: all
all: generate vendor build

.PHONY: generate
generate:
	rm -r ./pkg/languages/builders; \
	rm -r ./pkg/deployments/deployTypes; \
	GO111MODULE=on go generate ./pkg/languages/...; \
	GO111MODULE=on go generate ./pkg/deployments/...;


.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor;

.PHONY: build
build:
	GO111MODULE=on go build -v -o .

.PHONY: build-all
build-all: build-windows-amd64 build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -v -o ./bin/draftv2-windows-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -v -o ./bin/draftv2-linux-amd64

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -v -o ./bin/draftv2-linux-arm64

.PHONY: build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -v -o ./bin/draftv2-darwin-amd64

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -v -o ./bin/draft-v2darwin-arm64