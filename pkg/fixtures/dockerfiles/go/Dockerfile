FROM golang:1.23
ENV PORT=80
EXPOSE 80

WORKDIR /go/src/app
COPY . .

ARG GO111MODULE=off
RUN go build -v -o app ./main.go
RUN mv ./app /go/bin/

CMD ["app"]