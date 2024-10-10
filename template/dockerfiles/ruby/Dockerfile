FROM ruby:{{ .Config.GetVariableValue "VERSION" }}
ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}
RUN bundle config --global frozen 1

WORKDIR /usr/src/app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .
CMD ["ruby", "app.rb"]
