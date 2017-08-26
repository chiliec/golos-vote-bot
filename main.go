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

var (
	postingKey string
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"
)

func init() {
	flag.StringVar(&postingKey, "postingKey", "", "posting key")
	flag.Parse()
}

func main() {
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
		if update.Message != nil {
			userMessageText := update.Message.Text
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

			regexp, err := regexp.Compile("https://golos.io/([-a-zA-Z0-9@:%_+.~#?&//=]{2,256})/@([-a-zA-Z0-9]{2,256})/([-a-zA-Z0-9@:%_+.~#?&=]{2,256})")
			if err != nil {
				log.Panic(err)
			}
			if regexp.MatchString(userMessageText) {
				matched := regexp.FindStringSubmatch(userMessageText)
				log.Println(matched)
				author, permalink := matched[2], matched[3]
				voter := "chiliec"
				percent := 5
				vote(voter, author, permalink, percent)
				msg.ReplyToMessageID = update.Message.MessageID
				msg.Text = fmt.Sprintf("Проголосовал с силой %d%%", percent)
			}
			bot.Send(msg)
		}
	}
}

func vote(voter string, author string, permalink string, percent int) {
	if percent > 100 {
		percent = 100
	}
	weight := percent * 100
	client.Key_List = map[string]client.Keys{voter: client.Keys{postingKey, "", "", ""}}
	api := client.NewApi(rpc, chain)
	fmt.Println(api.Vote(voter, author, permalink, weight))
}
