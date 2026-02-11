FROM swift:5.5 AS build

WORKDIR /src
COPY . /src
RUN apt-get update \
  && apt-get install -y --no-install-recommends sudo openssl libssl-dev libcurl4-openssl-dev \
  && rm -rf /var/lib/apt/lists/*
RUN swift build -c release --product app

FROM swift:5.5

WORKDIR /app
COPY --from=build /src/.build/release/app ./app

ENV PORT 80
EXPOSE 80

CMD ["./app"]
