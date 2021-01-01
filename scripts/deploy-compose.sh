#!/bin/bash

# build the app and image
#docker build -t gymlog_go .

# deploy docker stack
docker-compose rm --stop --force -v
docker-compose -f docker-compose.yml up --build