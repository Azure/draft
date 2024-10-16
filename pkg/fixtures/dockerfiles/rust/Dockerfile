FROM rust:1.70.0

WORKDIR /usr/src/app
COPY . /usr/src/app
RUN cargo build

ENV PORT 80
EXPOSE 80

CMD ["cargo", "run", "-q"]
