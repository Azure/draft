FROM golang:1.18-alpine

WORKDIR /draftv2
COPY . ./

RUN apk add build-base

RUN GO111MODULE=on go generate ./pkg/languages/...
RUN GO111MODULE=on go generate ./pkg/deployments/...

RUN go mod vendor
CMD ["go", "test", "./...", "-buildvcs=false"]