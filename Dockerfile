FROM golang:latest as builder
RUN mkdir -p /go/src/github.com/GolosTools/golos-vote-bot
WORKDIR /go/src/github.com/GolosTools/golos-vote-bot
COPY . .
RUN GOOS=linux go build -a --ldflags '-extldflags "-static"' -o bin/golos-vote-bot -i .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder [ \
    "/go/src/github.com/GolosTools/golos-vote-bot/bin/golos-vote-bot", \
    "/go/src/github.com/GolosTools/golos-vote-bot/config.json", \
    "/go/src/github.com/GolosTools/golos-vote-bot/config.local.json", \
    "./"]
ENTRYPOINT ["./golos-vote-bot"]
