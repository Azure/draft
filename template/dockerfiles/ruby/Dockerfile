FROM ruby:{{VERSION}}
ENV PORT {{PORT}}
EXPOSE {{PORT}}
RUN bundle config --global frozen 1

WORKDIR /usr/src/app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .
CMD ["ruby", "app.rb"]
