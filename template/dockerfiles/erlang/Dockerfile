FROM erlang:{{ .Config.GetVariableValue "BUILDERVERSION" }} as builder

RUN apk add --update tar curl git bash make libc-dev gcc g++ && \
    rm -rf /var/cache/apk/*

RUN set -xe \
    && curl -fSL -o rebar3 "https://s3.amazonaws.com/rebar3/rebar3" \
    && chmod +x ./rebar3 \
    && ./rebar3 local install \
    && rm ./rebar3

WORKDIR /usr/src/app
COPY . /usr/src/app

ENV PATH "$PATH:/root/.cache/rebar3/bin"
RUN rebar3 as prod tar

RUN mkdir -p /opt/rel
RUN tar -zxvf /usr/src/app/_build/prod/rel/*/*.tar.gz -C /opt/rel

RUN relname=$(ls _build/prod/rel) ; echo $relname > /opt/rel/__relname

FROM alpine:{{ .Config.GetVariableValue "VERSION" }}

RUN apk add --no-cache openssl-dev ncurses libstdc++ libgcc

WORKDIR /opt/rel

ENV RELX_REPLACE_OS_VARS true
ENV HTTP_PORT {{ .Config.GetVariableValue "PORT" }}

COPY --from=builder /opt/rel /opt/rel

EXPOSE {{ .Config.GetVariableValue "PORT" }} {{ .Config.GetVariableValue "PORT" }}

RUN ln -s /opt/rel/bin/$(cat /opt/rel/__relname) /opt/rel/bin/start_script
ENTRYPOINT ["/opt/rel/bin/start_script"]

CMD ["foreground"]

