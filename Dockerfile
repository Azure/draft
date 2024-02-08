FROM golang:1.21-alpine

WORKDIR /draft

RUN apk add build-base
RUN apk add py3-pip
RUN apk add python3-dev libffi-dev openssl-dev cargo
RUN python3 -m venv .venv
RUN source .venv/bin/activate
RUN python3 -m pip install azure-cli
RUN apk add github-cli

COPY . ./
RUN make go-generate

RUN go mod vendor
ENTRYPOINT ["go"]