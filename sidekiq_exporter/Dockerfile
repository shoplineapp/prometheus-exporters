FROM ruby:2.4.2-slim-stretch

ENV SIDEKIQ_REDIS_ENDPOINT=""
ENV SIDEKIQ_REDIS_DATABASE=""

ENV APP_HOME /app

WORKDIR $APP_HOME

COPY config.ru Gemfile Gemfile.lock ./

RUN bundle install

EXPOSE 9292

CMD ["bundle","exec","rackup", "-p9292", "-o0.0.0.0"]