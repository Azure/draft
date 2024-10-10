FROM swift:{{ .Config.GetVariableValue "VERSION" }}

WORKDIR /src
COPY . /src
RUN apt-get update && apt-get install -y sudo openssl libssl-dev libcurl4-openssl-dev
RUN swift build -c release

ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

CMD ["swift", "run"]
