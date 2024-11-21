FROM golang:{{ .Config.GetVariableValue "VERSION" }}
ENV PORT={{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

WORKDIR /go/src/app
COPY . .

ARG GO111MODULE=off
RUN go build -v -o app ./main.go
RUN mv ./app /go/bin/

CMD ["app"]