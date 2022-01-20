FROM rust:1.19.0

WORKDIR /usr/src/app
COPY . /usr/src/app
RUN cargo install --path . && cargo build

ENV PORT {{PORT}}
EXPOSE {{PORT}}

CMD ["cargo", "run", "-q"]
