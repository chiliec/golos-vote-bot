# Голосовалочка
[![Build Status](https://travis-ci.org/GolosTools/golos-vote-bot.svg?branch=master)](https://travis-ci.org/GolosTools/golos-vote-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/GolosTools/golos-vote-bot)](https://goreportcard.com/report/github.com/GolosTools/golos-vote-bot)
[![Test Coverage](https://codeclimate.com/github/GolosTools/golos-vote-bot/badges/coverage.svg)](https://codeclimate.com/github/GolosTools/golos-vote-bot/coverage)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/GolosTools/golos-vote-bot/master/LICENSE)

Бот для коллективного кураторства на Голосе.

## Запуск

#### Шаг 1
Форкните репозиторий и склонируйте его себе через `go get`

#### Шаг 2
Скопируйте стандартный конфиг с новым именем `config.local.json`: 
```bash
cp config.json config.local.json
```
и измените нужные параметры в нём.

#### Шаг 3
Выполните:
```bash
go run main.go
```

#### Шаг 4
**Profit!**
