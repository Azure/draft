.PHONY: all
all: generate vendor build

.PHONY: generate
generate:
	rm -r ./pkg/languages/builders; \
	rm -r ./pkg/deployments/deployTypes; \
	go generate ./pkg/languages/...; \
	go generate ./pkg/deployments/...;


.PHONY: vendor
vendor:
	go mod tidy && go mod vendor;

.PHONY: build
build:
	go build -o .
