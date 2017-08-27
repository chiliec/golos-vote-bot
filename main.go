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
	logins     map[int]string
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"

	keyButtonText   = "🔑 Ключница"
	aboutButtonText = "🐞 О боте"
)

func init() {
	flag.StringVar(&postingKey, "postingKey", "", "posting key")
	flag.Parse()

	database = db.InitDB("./db/database.db")
	logins = map[int]string{}
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
	if err != nil {
		log.Panic(err)
	}
	for update := range updates {
		err := processMessage(bot, update)
		if err != nil {
			log.Println(err)
		}
	}
}

func processMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	log.Printf("[%s] %s", update.Message.From.UserName, "")
	if update.Message != nil {
		regexp, err := regexp.Compile("https://golos.io/([-a-zA-Z0-9@:%_+.~#?&//=]{2,256})/@([-a-zA-Z0-9]{2,256})/([-a-zA-Z0-9@:%_+.~#?&=]{2,256})")
		if err != nil {
			return err
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				keyButton := tgbotapi.NewKeyboardButton(keyButtonText)
				aboutButton := tgbotapi.NewKeyboardButton(aboutButtonText)
				buttons := []tgbotapi.KeyboardButton{keyButton, aboutButton}
				keyboard := tgbotapi.NewReplyKeyboard(buttons)
				msg.ReplyMarkup = keyboard
			}
		} else if update.Message.Text == keyButtonText {
			msg.Text = "Введите логин на Голосе"
			setWaitLogin(update.Message.From.ID)
		} else if update.Message.Text == aboutButtonText {
			msg.Text = "Бот для блого-социальной сети на блокчейне \"Голос\"\n" +
				"Нет времени голосовать, но хочется зарабатывать? Добавьте приватный постинг ключ и мы распорядимся вашей Силой голоса наилучшим образом!\n" +
				"Автор: @babin"
		} else if regexp.MatchString(update.Message.Text) {
			matched := regexp.FindStringSubmatch(update.Message.Text)
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
		} else if wait, login := isWaitingKey(update.Message.From.ID); wait && login == "" {
			msg.Text = "Введите приватный ключ"
			setWaitKey(update.Message.From.ID, update.Message.Text)
		} else if wait, login := isWaitingKey(update.Message.From.ID); wait && login != "" {
			log.Println("Сейчас нужно сохранить логин и приватный ключ!")
			//credential := models.Credential{
			//	UserID:     update.Message.From.ID,
			//	UserName:   login,
			//	PostingKey: update.Message.Text,
			//}
			//log.Println(credential)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.Text = "Логин и приватный ключ успешно сохранён!"
			forgetLogin(update.Message.From.ID)
		} else {
			msg.Text = "Команда не распознана"
		}
		bot.Send(msg)
	}
	return nil
}

func forgetLogin(userID int) {
	delete(logins, userID)
}

func setWaitLogin(userID int) {
	logins[userID] = ""
}

func setWaitKey(userID int, login string) {
	logins[userID] = login
}

func isWaitingKey(userID int) (bool, string) {
	for id, login := range logins {
		if userID == id {
			return true, login
		}
	}
	return false, ""
}

func vote(model models.Vote) error {
	exists := model.Exists(database)
	if exists {
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
