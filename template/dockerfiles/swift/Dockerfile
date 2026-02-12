FROM swift:{{ .Config.GetVariableValue "VERSION" }} AS build

WORKDIR /src
COPY Package.* /src/
RUN apt-get update \
  && apt-get install -y --no-install-recommends sudo openssl libssl-dev libcurl4-openssl-dev \
  && rm -rf /var/lib/apt/lists/*
RUN swift package resolve
COPY . /src
RUN swift build -c release \
  && bin_path=$(swift build -c release --show-bin-path) \
  && exe_list=$(find "$bin_path" -maxdepth 1 -type f -perm -u=x) \
  && exe_count=$(printf "%s\n" "$exe_list" | sed '/^$/d' | wc -l | tr -d ' ') \
  && if [ "$exe_count" -ne 1 ]; then \
       echo "Expected exactly 1 executable, found $exe_count:" >&2; \
       printf "%s\n" "$exe_list" >&2; \
       exit 1; \
     fi \
  && exe_path=$(printf "%s\n" "$exe_list" | head -n1) \
  && mkdir -p /app \
  && cp "$exe_path" /app/app

FROM swift:{{ .Config.GetVariableValue "VERSION" }}

WORKDIR /app
COPY --from=build /app/app ./app

ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}

CMD ["./app"]
