FROM golang:{{VERSION}} AS builder
ENV PORT {{PORT}}
EXPOSE {{PORT}}

WORKDIR /go/src/app
COPY . .

RUN go get
RUN go build -v -o app
RUN mv ./app /go/bin/

FROM alpine:latest
COPY --from=builder /go/bin/app /usr/local/bin/app
CMD ["/usr/local/bin/app"]
