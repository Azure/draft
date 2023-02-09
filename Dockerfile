FROM golang:1.18-alpine

WORKDIR /draft

RUN apk add build-base
RUN apk add py3-pip
RUN apk add python3-dev libffi-dev openssl-dev cargo
RUN pip install --upgrade pip
RUN pip install azure-cli
RUN apk add github-cli

COPY . ./
RUN make go-generate

RUN go mod vendor
ENTRYPOINT ["go"]