FROM ubuntu

RUN mkdir /app
ADD bin/cache /usr/local/bin/cache
ADD ./config.yaml /app/
WORKDIR /app

ENTRYPOINT /usr/local/bin/cache -redis=$REDIS
