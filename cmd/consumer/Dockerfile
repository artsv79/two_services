FROM ubuntu

RUN mkdir /app
ADD bin/consumer /usr/local/bin/consumer
WORKDIR /app

ENTRYPOINT /usr/local/bin/consumer -cache=$CACHE
