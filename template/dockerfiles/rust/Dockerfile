FROM rust:{{ .Config.GetVariableValue "VERSION" }}

WORKDIR /usr/src/app
COPY . /usr/src/app
RUN cargo build

ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

CMD ["cargo", "run", "-q"]
