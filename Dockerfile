FROM ubuntu:17.04
WORKDIR /opt/golosbot/
COPY bin/golos-vote-bot .
COPY config.json .
RUN apt-get update && apt-get install -y ca-certificates
ENTRYPOINT ./golos-vote-bot
