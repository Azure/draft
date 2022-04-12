FROM golang:1.18-alpine

WORKDIR /draftv2
COPY . ./

RUN apk add build-base

RUN make go-generate

RUN go mod vendor
CMD ["go", "test", "./...", "-buildvcs=false"]