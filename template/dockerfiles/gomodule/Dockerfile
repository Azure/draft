FROM golang:{{ .Config.GetVariableValue "VERSION" }} AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app-binary

FROM gcr.io/distroless/static-debian12

ENV PORT={{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

WORKDIR /app
COPY --from=builder /build/app-binary . 
CMD ["/app/app-binary"]
