package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/Chiliec/golos-go/client"
	"gopkg.in/telegram-bot-api.v4"

	"github.com/Chiliec/golos-vote-bot/db"
	"github.com/Chiliec/golos-vote-bot/models"
)

var (
	postingKey string
	database   *sql.DB
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"
)

func init() {
	flag.StringVar(&postingKey, "postingKey", "", "posting key")
	flag.Parse()

	database = db.InitDB("./db/database.db")
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
				percent := 65
				voteModel := models.Vote{
					UserID:    update.Message.From.ID,
					Voter:     voter,
					Author:    author,
					Permalink: permalink,
					Percent:   percent,
				}
				msg.ReplyToMessageID = update.Message.MessageID
				err := vote(voteModel)
				if err != nil {
					switch err.(type) {
					case *ErrorAlreadyVoted:
						msg.Text = "Уже голосовал за этот пост!"
					default:
						msg.Text = "Не смог прогосовать, попробуйте ещё раз"
					}
				} else {
					msg.Text = fmt.Sprintf("Проголосовал с силой %d%%", percent)
				}
			}
			bot.Send(msg)
		}
	}
}

func vote(model models.Vote) error {
	exists := model.Exists(database)
	if exists {
		log.Println("Already voted!!!")
		return NewErrorAlreadyVoted("Уже проголосовали!")
	}
	weight := model.Percent * 100
	client.Key_List = map[string]client.Keys{model.Voter: client.Keys{postingKey, "", "", ""}}
	api := client.NewApi(rpc, chain)
	err := api.Vote(model.Voter, model.Author, model.Permalink, weight)
	if err != nil {
		return err
	}
	_, err = model.Save(database)
	if err != nil {
		return err
	}
	return nil
}

type ErrorAlreadyVoted struct {
	message string
}

func NewErrorAlreadyVoted(message string) *ErrorAlreadyVoted {
	return &ErrorAlreadyVoted{
		message: message,
	}
}
func (e *ErrorAlreadyVoted) Error() string {
	return e.message
}
