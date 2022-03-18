FROM golang:1.18-alpine

WORKDIR /draftv2
COPY . ./

RUN apk add build-base
RUN go mod vendor
RUN go test ./... -buildvcs=false