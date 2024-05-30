FROM golang:1.22-alpine

WORKDIR /draft

RUN apk add gcc musl-dev libffi-dev openssl-dev cargo make wget zlib-dev

ARG PYTHON_VERSION=3.9.9
RUN cd /opt \
    && wget https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \                                              
    && tar xzf Python-${PYTHON_VERSION}.tgz
RUN cd /opt/Python-${PYTHON_VERSION} \ 
    && ./configure --prefix=/usr --enable-optimizations --with-ensurepip=install \
    && make install \
    && rm /opt/Python-${PYTHON_VERSION}.tgz /opt/Python-${PYTHON_VERSION} -rf

RUN python3 -m venv az-cli-env
RUN az-cli-env/bin/pip install --upgrade pip
RUN az-cli-env/bin/pip install --upgrade azure-cli
RUN az-cli-env/bin/az --version

ENV PATH "$PATH:/draft/az-cli-env/bin"

RUN apk add github-cli

COPY ./go.mod ./go.mod
RUN go mod download

COPY . ./
RUN make go-generate

ENTRYPOINT ["go"]
