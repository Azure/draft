FROM rust:1.50.0

WORKDIR /usr/src/app
COPY . /usr/src/app
RUN cargo install && cargo build

ENV PORT {{PORT}}
EXPOSE {{PORT}}

CMD ["cargo", "run", "-q"]
