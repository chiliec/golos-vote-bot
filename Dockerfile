FROM ubuntu:17.04

COPY ./bin/golos-vote-bot /opt/golosbot/

WORKDIR /opt/golosbot/

RUN apt-get update && apt-get install -y ca-certificates
