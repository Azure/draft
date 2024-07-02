FROM golang:1.18
ENV PORT 8080
EXPOSE 8080

WORKDIR /go/src/app
COPY . .

ARG GO111MODULE=off
RUN go build -v -o app ./main.go
RUN mv ./app /go/bin/

CMD ["app"]