FROM rust:1.58.0

WORKDIR /usr/src/app
COPY . /usr/src/app
RUN cargo build

ENV PORT {{PORT}}
EXPOSE {{PORT}}

CMD ["cargo", "run", "-q"]
