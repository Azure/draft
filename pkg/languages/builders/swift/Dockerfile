FROM swift:{{VERSION}}

WORKDIR /src
COPY . /src
RUN apt-get update && apt-get install -y sudo openssl libssl-dev libcurl4-openssl-dev
RUN swift build -c release

ENV PORT {{PORT}}
EXPOSE {{PORT}}

CMD ["swift", "run"]
