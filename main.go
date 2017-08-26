package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Chiliec/golos-go/client"
	"github.com/pkg/errors"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"
)

func main() {
	author, permlink := "anima", "konkurs-nochnye-ulicy-poslednie-raboty-stop-i-golosovanie"
	voter := "chiliec"
	weight := 10000
	var postingKey string
	flag.StringVar(&postingKey, "postingKey", "", "posting key")
	flag.Parse()
	client.Key_List = map[string]client.Keys{voter: client.Keys{postingKey, "", "", ""}}
	api := client.NewApi(rpc, chain)
	fmt.Println(api.Vote(voter, author, permlink, weight))

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
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
