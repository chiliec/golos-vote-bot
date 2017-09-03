#!/bin/bash

GOOS=linux go build -a --ldflags '-extldflags "-static"' -o bin/golos-vote-bot -i .
