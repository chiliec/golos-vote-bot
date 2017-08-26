package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/Chiliec/golos-go/client"
	"github.com/pkg/errors"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"
)

func main() {
	url := "https://golos.io/ru--zhiznx/@gothy/spirt-lotreamon-i-beshenye-psy-ili-kak-ya-ne-pogib-v-90-e"
	regexp, err := regexp.Compile("https://golos.io/([-a-zA-Z0-9@:%_+.~#?&//=]{2,256})/@([-a-zA-Z0-9]{2,256})/([-a-zA-Z0-9@:%_+.~#?&=]{2,256})")
	if err != nil {
		log.Panic(err)
	}
	matched := regexp.FindStringSubmatch(url)
	author, permalink := matched[2], matched[3]
	voter := "chiliec"
	weight := 10000
	var postingKey string
	flag.StringVar(&postingKey, "postingKey", "", "posting key")
	flag.Parse()
	client.Key_List = map[string]client.Keys{voter: client.Keys{postingKey, "", "", ""}}
	api := client.NewApi(rpc, chain)
	fmt.Println(api.Vote(voter, author, permalink, weight))

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Panic(errors.New("Нет токена"))
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				keyButton := tgbotapi.NewKeyboardButton("🔑 Ключница")
				aboutButton := tgbotapi.NewKeyboardButton("🐞 О боте")
				buttons := []tgbotapi.KeyboardButton{keyButton, aboutButton}
				keyboard := tgbotapi.NewReplyKeyboard(buttons)
				msg.ReplyMarkup = keyboard
			}
		}
		bot.Send(msg)
	}
}
