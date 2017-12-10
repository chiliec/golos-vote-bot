# Голосовалочка
[![Build Status](https://travis-ci.org/GolosTools/golos-vote-bot.svg?branch=master)](https://travis-ci.org/GolosTools/golos-vote-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/GolosTools/golos-vote-bot)](https://goreportcard.com/report/github.com/GolosTools/golos-vote-bot)
[![Test Coverage](https://codeclimate.com/github/GolosTools/golos-vote-bot/badges/coverage.svg)](https://codeclimate.com/github/GolosTools/golos-vote-bot/coverage)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/GolosTools/golos-vote-bot/master/LICENSE)

Бот для коллективного кураторства в [социальной блокчейн-сети "Голос"](https://ru.wikipedia.org/wiki/Голос_(социальная_сеть)).

## Запуск

### Шаг 1
Форкните репозиторий и склонируйте его через `go get`

### Шаг 2
Скопируйте стандартный конфиг с новым именем `config.local.json`: 
```bash
cp config.json config.local.json
```
и измените нужные параметры в нём.

### Шаг 3
Выполните:
```bash
go run main.go
```

## Деплой в Docker

Выполните команды:
```bash
docker build --no-cache -t golosovalochka .
docker stop golosovalochka
docker rm golosovalochka
docker run -d -v `pwd`/db:/root/db --name golosovalochka golosovalochka:latest .
```
или воспользуйтесь файлом `./redeploy.sh`

## Обновление зависимостей

Для управления зависимостями используется [Dep](https://github.com/golang/dep).

Для обновления зависимостей выполните команду `dep ensure -update`.

## Лицензия
Лицензия [MIT](https://github.com/GolosTools/golos-vote-bot/blob/master/LICENSE).
Свободно используйте, распространяйте и не забывайте контрибьютить обратно.
