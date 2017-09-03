package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/asuleymanov/golos-go/client"
	"gopkg.in/telegram-bot-api.v4"

	"github.com/Chiliec/golos-vote-bot/db"
	"github.com/Chiliec/golos-vote-bot/models"
)

var (
	database *sql.DB
	logins   map[int]string
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"

	keyButtonText   = "🔑 Ключница"
	aboutButtonText = "🐞 О боте"

	groupLink = "https://t.me/joinchat/AlKeQUQpN8-9oShtaTcY7Q"
	groupID   = -1001143551951

	waitMinutes = 1
)

var golos = client.NewApi(rpc, chain)

var alreadyVotedError = errors.New("Уже проголосовали!")

func init() {
	db, err := db.InitDB("./db/database.db")
	if err != nil {
		log.Panic(err)
	}
	database = db
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
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		return errors.New("Не получили ID чата")
	}
	msg := tgbotapi.NewMessage(chatID, "")
	if update.Message != nil {
		regexp, err := regexp.Compile("https://golos.io/([-a-zA-Z0-9@:%_+.~#?&//=]{2,256})/@([-a-zA-Z0-9.]{2,256})/([-a-zA-Z0-9@:%_+.~#?&=]{2,256})")
		if err != nil {
			return err
		}
		switch {
		case update.Message.IsCommand():
			switch update.Message.Command() {
			case "start":
				keyButton := tgbotapi.NewKeyboardButton(keyButtonText)
				aboutButton := tgbotapi.NewKeyboardButton(aboutButtonText)
				buttons := []tgbotapi.KeyboardButton{keyButton, aboutButton}
				keyboard := tgbotapi.NewReplyKeyboard(buttons)
				msg.ReplyMarkup = keyboard
			}
		case update.Message.Text == keyButtonText:
			msg.Text = "Введите логин на Голосе"
			setWaitLogin(update.Message.From.ID)
		case update.Message.Text == aboutButtonText:
			msg.Text = "Бот для блого-социальной сети на блокчейне \"Голос\"\n" +
				"Нет времени голосовать, но хочется зарабатывать? Добавьте приватный постинг ключ и мы распорядимся вашей Силой голоса наилучшим образом!\n" +
				"Автор: @babin"
			forgetLogin(update.Message.From.ID)
		case regexp.MatchString(update.Message.Text):
			matched := regexp.FindStringSubmatch(update.Message.Text)
			author, permalink := matched[2], matched[3]

			percent := 1
			if update.Message.Chat.Type != "private" {
				percent = 3
			}
			if update.Message.Chat.ID == groupID {
				percent = 100
			}

			msg.ReplyToMessageID = update.Message.MessageID

			voteModel := models.Vote{
				UserID:    update.Message.From.ID,
				Author:    author,
				Permalink: permalink,
				Percent:   percent,
			}

			if voteModel.Exists(database) {
				msg.Text = "Уже голосовал за этот пост!"
				break
			}

			_, err = voteModel.Save(database)
			if err != nil {
				log.Println("Error save vote model: " + err.Error())
				msg.Text = "Не смог проголосовать за пост"
				break
			}

			msg.Text = "Открываю голосование:"
			goodButton := tgbotapi.NewInlineKeyboardButtonData("Хороший пост", "post-id_good")
			badButton := tgbotapi.NewInlineKeyboardButtonData("Плохой пост", "post-id_bad")
			buttons := []tgbotapi.InlineKeyboardButton{}
			buttons = append(buttons, goodButton)
			row := []tgbotapi.InlineKeyboardButton{goodButton, badButton}
			markup := tgbotapi.InlineKeyboardMarkup{}
			markup.InlineKeyboard = append(markup.InlineKeyboard, row)
			msg.ReplyMarkup = markup

			go func(voteModel models.Vote, messageID int) {
				select {
				case <-time.After(time.Second * 60 * waitMinutes):
					successVotes := vote(voteModel)
					message := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					message.ReplyToMessageID = messageID
					message.Text = fmt.Sprintf("Проголосовал с силой %d%% c %d аккаунтов", percent, successVotes)
					bot.Send(message)
				}
			}(voteModel, update.Message.MessageID)
		default:
			if wait, login := isWaitingKey(update.Message.From.ID); wait {
				if login == "" {
					msg.Text = "Введите приватный ключ"
					setWaitKey(update.Message.From.ID, update.Message.Text)
				} else {
					credential := models.Credential{
						UserID:     update.Message.From.ID,
						UserName:   login,
						PostingKey: update.Message.Text,
					}

					client.Key_List[credential.UserName] = client.Keys{PKey: credential.PostingKey}
					// TODO: find method to just verify posting key without any actions
					if err := golos.Follow(credential.UserName, "chiliec"); err == nil {
						result, err := credential.Save(database)
						if err != nil {
							log.Println(err.Error())
						}
						if result {
							msg.Text = "Логин и приватный ключ успешно сохранён! " +
								"Присоединяйтесь к нашей группе: " + groupLink
						} else {
							msg.Text = "Не смог сохранить логин и приватный ключ :("
						}
					} else {
						log.Println("Не сохранили ключ: " + err.Error())
						msg.Text = "Логин и приватный ключ не совпадают :("
					}

					forgetLogin(update.Message.From.ID)
				}
			} else if update.Message.Chat.Type == "private" {
				msg.Text = "Не понимаю"
			}
		}
	} else if update.CallbackQuery != nil {
		log.Println(update.CallbackQuery)
	}
	if msg.Text != "" {
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

func vote(vote models.Vote) int {
	credentials, err := models.GetAllCredentials(database)
	if err != nil {
		log.Println("Не смогли извлечь ключи из базы")
		return 0
	}
	for _, credential := range credentials {
		client.Key_List[credential.UserName] = client.Keys{PKey: credential.PostingKey}
	}
	log.Printf("Загружено %d аккаунтов", len(credentials))

	var errors []error
	var wg sync.WaitGroup
	wg.Add(len(credentials))
	for _, credential := range credentials {
		client.Key_List[credential.UserName] = client.Keys{PKey: credential.PostingKey}
		go func(credential models.Credential) {
			defer wg.Done()
			weight := vote.Percent * 100
			err := golos.Vote(credential.UserName, vote.Author, vote.Permalink, weight)
			if err != nil {
				errors = append(errors, err)
			}
		}(credential)
	}
	wg.Wait()
	return len(credentials) - len(errors)
}
