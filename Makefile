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
	GO111MODULE=on go mod tidy; \
	GO111MODULE=on go mod vendor;

.PHONY: build
build:
	GO111MODULE=on go build -v -o .
