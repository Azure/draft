FROM swift:{{ .Config.GetVariableValue "VERSION" }} AS build

WORKDIR /src
COPY . /src
RUN apt-get update \
  && apt-get install -y --no-install-recommends sudo openssl libssl-dev libcurl4-openssl-dev \
  && rm -rf /var/lib/apt/lists/*
RUN swift build -c release --product {{ .Config.GetVariableValue "APPNAME" }}

FROM swift:{{ .Config.GetVariableValue "VERSION" }}

WORKDIR /app
COPY --from=build /src/.build/release/{{ .Config.GetVariableValue "APPNAME" }} ./{{ .Config.GetVariableValue "APPNAME" }}

ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

CMD ["./{{ .Config.GetVariableValue "APPNAME" }}"]
