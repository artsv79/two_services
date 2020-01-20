#!/bin/bash

go build -a -o bin/cache github.com/artsv79/2services/cmd/cache
go build -a -o bin/consumer github.com/artsv79/2services/cmd/consumer
docker build -t 2services-cache -f cmd/cache/Dockerfile .
docker build -t 2services-consumer -f cmd/consumer/Dockerfile .

