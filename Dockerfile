FROM golang:1.18-alpine

WORKDIR /draftv2
COPY . ./

RUN apk add build-base
RUN apk add py3-pip
RUN apk add gcc musl-dev python3-dev libffi-dev openssl-dev cargo make
RUN pip install --upgrade pip
RUN pip install azure-cli

RUN make go-generate

RUN go mod vendor
ENTRYPOINT ["go"]