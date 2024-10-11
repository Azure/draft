FROM swift:5.5

WORKDIR /src
COPY . /src
RUN apt-get update && apt-get install -y sudo openssl libssl-dev libcurl4-openssl-dev
RUN swift build -c release

ENV PORT 80
EXPOSE 80

CMD ["swift", "run"]
