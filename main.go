package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/asuleymanov/golos-go/client"
	"github.com/go-telegram-bot-api/telegram-bot-api"

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

	addKeyButtonText    = "🗝 Добавить ключ"
	removeKeyButtonText = "❌ Удалить ключ"

	groupLink = "https://t.me/joinchat/AlKeQUQpN8-9oShtaTcY7Q"
	groupID   = -1001143551951

	requiredVotes     = 2
	initialUserRating = 10
)

var alreadyVotedError = errors.New("уже проголосовали")

func init() {
	db, err := db.InitDB("./db/database.db")
	db.SetMaxOpenConns(1)
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

	bot.Debug = false

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
	chatID, err := getChatID(update)
	if err != nil {
		return err
	}
	userID, err := getUserID(update)
	if err != nil {
		return err
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
				if update.Message.Chat.Type != "private" {
					msg.Text = "Я такое только в личке буду обсуждать"
					break
				}
				addKeyButton := tgbotapi.NewKeyboardButton(addKeyButtonText)
				removeKeyButton := tgbotapi.NewKeyboardButton(removeKeyButtonText)
				firstButtonRow := []tgbotapi.KeyboardButton{addKeyButton, removeKeyButton}
				keyboard := tgbotapi.NewReplyKeyboard(firstButtonRow)
				msg.ReplyMarkup = keyboard
				msg.Text = fmt.Sprintf("Привет, %s! \n\n"+
					"Я — бот для коллективного кураторства в [социальной блокчейн-сети \"Голос\"](https://golos.io).\n\n"+
					"Предлагаю начать с добавления приватного постинг-ключа нажатием кнопки \""+addKeyButtonText+"\""+
					", после чего я дам ссылку на группу куда предлагать посты для поддержки.\n\n"+
					"По любым вопросам пиши моему хозяину — @babin",
					update.Message.From.FirstName)
				forgetLogin(userID)
			}
		case update.Message.Text == addKeyButtonText:
			if update.Message.Chat.Type != "private" {
				msg.Text = "Я такое только в личке буду обсуждать"
				break
			}
			msg.Text = "Введи логин на Голосе"
			setWaitLogin(userID)
		case update.Message.Text == removeKeyButtonText:
			if update.Message.Chat.Type != "private" {
				msg.Text = "Я такое только в личке буду обсуждать"
				break
			}
			credential, err := models.GetCredentialByUserID(userID, database)
			msg.Text = "Произошла ошибка при удалении ключа"
			if err == nil {
				credential.PostingKey = ""
				result, err := credential.Save(database)
				if result && err == nil {
					msg.Text = "Твой ключ успешно удалён. Я больше не буду отвечать на твои предложения по курированию постов."
				}
			}
			forgetLogin(userID)
		case regexp.MatchString(update.Message.Text):
			msg.ReplyToMessageID = update.Message.MessageID

			if update.Message.Chat.Type == "private" {
				msg.Text = "Присоединяйтесь к нашей группе: " + groupLink
				break
			}

			credential, err := models.GetCredentialByUserID(userID, database)
			if err != nil || len(credential.PostingKey) == 0 {
				msg.Text = "Не могу допустить тебя к кураторству, у меня ещё нет твоего постинг-ключа. Напиши мне в личку, обсудим этот вопрос"
				break
			}

			matched := regexp.FindStringSubmatch(update.Message.Text)
			author, permalink := matched[2], matched[3]

			percent := 5
			if chatID == groupID {
				percent = 100
			}

			voteModel := models.Vote{
				UserID:    userID,
				Author:    author,
				Permalink: permalink,
				Percent:   percent,
			}

			if voteModel.Exists(database) {
				msg.Text = "Уже голосовал за этот пост!"
				break
			}

			voteID, err := voteModel.Save(database)
			if err != nil {
				return err
			}

			msg.Text = "Голосование открыто"
			markup := getVoteMarkup(voteID, 0, 0)
			msg.ReplyMarkup = markup
		default:
			if update.Message.Chat.Type != "private" {
				return nil
			}
			msg.Text = "Не понимаю"
			if wait, login := isWaitingKey(userID); wait {
				if login == "" {
					login = strings.Trim(update.Message.Text, "@")
					setWaitKey(userID, login)
					msg.Text = "Введите приватный ключ"
				} else {
					credential := models.Credential{
						UserID:     userID,
						UserName:   login,
						PostingKey: update.Message.Text,
						Rating:     initialUserRating,
					}
					if rating, err := credential.GetRating(database); err == nil {
						credential.Rating = rating
					}

					golos := client.NewApi([]string{rpc}, chain)
					defer golos.Rpc.Close()
					if golos.Login(credential.UserName, credential.PostingKey) {
						result, err := credential.Save(database)
						if err != nil {
							log.Println(err.Error())
						}
						if result {
							msg.Text = "Логин и приватный ключ успешно сохранён! " +
								"Присоединяйтесь к нашей группе для участия в курировании: " + groupLink
						} else {
							msg.Text = "Не смог сохранить логин и приватный ключ :("
						}
					} else {
						msg.Text = "Логин и приватный ключ не совпадают :("
					}

					forgetLogin(userID)
				}
			}
		}
	} else if update.CallbackQuery != nil {
		log.Println(update.CallbackQuery.Message)
		arr := strings.Split(update.CallbackQuery.Data, "_")
		voteStringID := arr[0]
		action := arr[1]
		voteID, err := strconv.ParseInt(voteStringID, 10, 64)
		if err != nil {
			return err
		}
		isGood := action == "good"
		response := models.Response{
			UserID: userID,
			VoteID: voteID,
			Result: isGood,
		}
		text := "Вы уже голосовали!"
		responseExists := response.Exists(database)
		if !responseExists {
			text = "Голос принят"
		}
		config := tgbotapi.CallbackConfig{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            text,
			ShowAlert:       true,
		}
		bot.AnswerCallbackQuery(config)

		if !responseExists {
			_, err := response.Save(database)
			if err != nil {
				return err
			}
			voteModel := models.GetVote(database, voteID)
			err = verifyVotes(bot, voteModel, update)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if msg.Text == "" {
		return errors.New("отсутствует текст сообщения")
	}
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	bot.Send(msg)
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

func verifyVotes(bot *tgbotapi.BotAPI, voteModel models.Vote, update tgbotapi.Update) error {
	chatID, err := getChatID(update)
	if err != nil {
		return err
	}
	userID, err := getUserID(update)
	if err != nil {
		return err
	}
	messageID, err := getMessageID(update)
	if err != nil {
		return err
	}

	responses, err := models.GetAllResponsesForVoteID(voteModel.VoteID, database)
	if err != nil {
		return err
	}

	var positives, negatives int
	for _, response := range responses {
		if response.Result {
			positives = positives + 1
		} else {
			negatives = negatives + 1
		}
	}

	markup := getVoteMarkup(voteModel.VoteID, positives, negatives)
	updateTextConfig := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   messageID,
			ReplyMarkup: &markup,
		},
		Text: update.CallbackQuery.Message.Text,
	}
	bot.Send(updateTextConfig)

	credential := models.Credential{UserID: userID}

	if positives+negatives >= requiredVotes {
		msg := tgbotapi.NewEditMessageText(chatID, messageID, "")
		if positives >= negatives {
			credential.IncrementRating(database)

			successVotes := vote(voteModel)
			msg.Text = fmt.Sprintf("Проголосовал с силой %d%% c %d аккаунтов", voteModel.Percent, successVotes)
		} else {
			credential.DecrementRating(database)
			rating, err := credential.GetRating(database)
			if err != nil {
				return err
			}
			msg.Text = "Пост отклонен, рейтинг предлагающего снижен"
			if rating < 0 {
				memberConfig := tgbotapi.KickChatMemberConfig{
					ChatMemberConfig: tgbotapi.ChatMemberConfig{
						ChatID: chatID,
						UserID: userID,
					},
					UntilDate: 0,
				}
				bot.KickChatMember(memberConfig)
				msg.Text = "Пост отклонен, предлагающий исключен"
			}
		}
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
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
			golos := client.NewApi([]string{rpc}, chain)
			defer golos.Rpc.Close()
			err := golos.Vote(credential.UserName, vote.Author, vote.Permalink, weight)
			if err != nil {
				log.Println("Ошибка при голосовании: " + err.Error())
				errors = append(errors, err)
			}
		}(credential)
	}
	wg.Wait()
	return len(credentials) - len(errors)
}

func getVoteMarkup(voteID int64, positives int, negatives int) tgbotapi.InlineKeyboardMarkup {
	stringVoteID := strconv.FormatInt(voteID, 10)
	goodButton := tgbotapi.NewInlineKeyboardButtonData("👍 Хороший пост ("+strconv.Itoa(positives)+")", stringVoteID+"_good")
	badButton := tgbotapi.NewInlineKeyboardButtonData("👎 Плохой пост ("+strconv.Itoa(negatives)+")", stringVoteID+"_bad")
	buttons := []tgbotapi.InlineKeyboardButton{}
	buttons = append(buttons, goodButton)
	row := []tgbotapi.InlineKeyboardButton{goodButton, badButton}
	markup := tgbotapi.InlineKeyboardMarkup{}
	markup.InlineKeyboard = append(markup.InlineKeyboard, row)
	return markup
}

func getChatID(update tgbotapi.Update) (int64, error) {
	if update.Message != nil {
		return update.Message.Chat.ID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID, nil
	} else {
		return 0, errors.New("не получили ID чата")
	}
}

func getUserID(update tgbotapi.Update) (int, error) {
	if update.Message != nil {
		return update.Message.From.ID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID, nil
	} else {
		return 0, errors.New("не получили ID пользователя")
	}
}

func getMessageID(update tgbotapi.Update) (int, error) {
	if update.Message != nil {
		return update.Message.MessageID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.MessageID, nil
	} else {
		return 0, errors.New("не получили ID сообщения")
	}
}
