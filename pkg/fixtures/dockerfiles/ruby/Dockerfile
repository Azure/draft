FROM ruby:3.1.2
ENV PORT 80
EXPOSE 80
RUN bundle config --global frozen 1

WORKDIR /usr/src/app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .
CMD ["ruby", "app.rb"]
