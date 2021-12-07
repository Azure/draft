FROM golang
ENV PORT {{PORT}}
EXPOSE {{PORT}}

WORKDIR /go/src/app
COPY . .

RUN go mod vendor
RUN go build -v -o app
RUN mv ./app /go/bin/

CMD ["app"]