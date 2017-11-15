FROM golang:1.9.2 as builder
RUN mkdir -p /go/src/github.com/GolosTools/golos-vote-bot
WORKDIR /go/src/github.com/GolosTools/golos-vote-bot
COPY . .
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep \
    && dep ensure -vendor-only
RUN GOOS=linux go build -a --ldflags '-extldflags "-static"' -o bin/golos-vote-bot -i .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder [ \
    "/go/src/github.com/GolosTools/golos-vote-bot/bin/golos-vote-bot", \
    "/go/src/github.com/GolosTools/golos-vote-bot/config.json", \
    "/go/src/github.com/GolosTools/golos-vote-bot/config.local.json", \
    "/root/"]
ENTRYPOINT ["./golos-vote-bot"]
