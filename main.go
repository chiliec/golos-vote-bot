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
	"time"

	"github.com/asuleymanov/golos-go/client"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/GolosTools/golos-vote-bot/config"
	"github.com/GolosTools/golos-vote-bot/db"
	"github.com/GolosTools/golos-vote-bot/helpers"
	"github.com/GolosTools/golos-vote-bot/models"
)

const (
	addKeyButtonText    = "🗝Добавить ключ"
	removeKeyButtonText = "❌Удалить ключ"
)

func main() {
	var configuration config.Config
	err := config.LoadConfiguration("./config.json", &configuration)
	if err != nil {
		log.Panic(err)
	}
	err = config.LoadConfiguration("./config.local.json", &configuration)
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}

	database, err := db.InitDB(configuration.DatabasePath)
	if err != nil {
		log.Panic(err)
	}
	defer database.Close()

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramToken)
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
		err := processMessage(bot, update, configuration, database)
		if err != nil {
			log.Println(err)
		}
	}
}

func processMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, config config.Config, database *sql.DB) error {
	chatID, err := getChatID(update)
	if err != nil {
		return err
	}
	userID, err := getUserID(update)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatID, "")

	state, err := models.GetStateByUserID(userID, database)
	if err != nil {
		return err
	}

	if update.Message != nil {
		regexp, err := regexp.Compile("https://(?:golos.io|goldvoice.club)(?:[-a-zA-Z0-9@:%_+.~#?&//=]{2,256})?/@([-a-zA-Z0-9.]{2,256})/([-a-zA-Z0-9@:%_+.~?&=]{2,256})")
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
					"Мой код полностью открыт и находится здесь: https://github.com/GolosTools/golos-vote-bot\n\n"+
					"Предлагаю начать с добавления ключа нажатием кнопки \""+addKeyButtonText+"\", "+
					"после чего я дам ссылку на группу для предложения постов.\n\n"+
					"По любым вопросам пиши моему хозяину — %s",
					update.Message.From.FirstName, config.Developer)

			}
			state.Action = update.Message.Command()
		case update.Message.Text == addKeyButtonText:
			if update.Message.Chat.Type != "private" {
				msg.Text = "Я такое только в личке буду обсуждать"
				break
			}
			msg.Text = fmt.Sprintf("Добавь доверенный аккаунт *%s* в https://golos.cf/multi/, "+
				"а затем скажи мне свой логин на Голосе", config.Account)
			state.Action = addKeyButtonText
		case update.Message.Text == removeKeyButtonText:
			if update.Message.Chat.Type != "private" {
				msg.Text = "Я такое только в личке буду обсуждать"
				break
			}
			msg.Text = "Произошла ошибка при удалении ключа"
			credential, err := models.GetCredentialByUserID(userID, database)
			if err == nil {
				if len(credential.UserName) == 0 || false == credential.Active {
					msg.Text = "У тебя ещё нет моего ключа. " +
						"Жми кнопку " + addKeyButtonText + "для добавления или используй команду " +
						"/start для обновления списка кнопок."
					break
				}
				credential.Active = false
				result, err := credential.Save(database)
				if true == result && err == nil {
					msg.Text = "Успех. Я больше не буду использовать твой аккаунт при курировании постов. " +
						"Дополнительно можешь удалить все сторонние ключи из своего аккаунта здесь: " +
						"https://golos.cf/multi/off.html"
				}
			}
			state.Action = removeKeyButtonText
		case regexp.MatchString(update.Message.Text):
			msg.ReplyToMessageID = update.Message.MessageID

			matched := regexp.FindStringSubmatch(update.Message.Text)
			author, permalink := matched[1], matched[2]

			golos := client.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			post, err := golos.Rpc.Database.GetContent(author, permalink)
			if err != nil {
				return err
			}
			// check post exists in blockchain
			if post.Author != author || post.Permlink != permalink {
				return nil
			}

			if update.Message.Chat.ID != config.GroupID {
				msg.Text = "Удобный просмотр с мобильных устройств:\n" + getInstantViewLink(author, permalink)
				msg.DisableWebPagePreview = false
				bot.Send(msg)
				return nil
			}

			if update.Message.Chat.Type == "private" {
				msg.Text = "Предложить пост можно в нашей группе " + config.GroupLink
				break
			}

			if models.GetTodayVotesCountForUserID(userID, database) >= config.MaximumUserVotesPerDay {
				msg.Text = "Лимит твоих постов на сегодня превышен. Приходи завтра!"
				break
			}

			if models.GetLastVote(database).UserID == userID {
				msg.Text = "Нельзя предлагать два поста подряд. Наберись терпения!"
				break
			}

			if models.GetOpenedVotesCount(database) >= config.MaximumOpenedVotes {
				msg.Text = "Слишком много уже открытых голосований. Может сначала с ними разберёмся?"
				break
			}

			credential, err := models.GetCredentialByUserID(userID, database)
			if err != nil {
				return err
			}
			if false == credential.Active {
				msg.Text = "Не могу допустить тебя к кураторству, у тебя ещё нет моего ключа. " +
					"Напиши мне в личку, обсудим этот вопрос"
				break
			}

			if post.Mode != "first_payout" {
				msg.Text = "Выплата за пост уже была произведена!"
				break
			}

			if post.MaxAcceptedPayout == "0.000 GBG" {
				msg.Text = "Мне не интересно голосовать за пост с отключенными выплатами"
				break
			}

			if helpers.IsVoxPopuli(author) {
				msg.Text = "Сообщества vox-populi могут сами себя поддержать"
				break
			}

			percent := 100

			voteModel := models.Vote{
				UserID:    userID,
				Author:    author,
				Permalink: permalink,
				Percent:   percent,
				Date:      time.Now(),
			}

			if voteModel.Exists(database) {
				msg.Text = "Уже голосовала за этот пост!"
				break
			}

			voteID, err := voteModel.Save(database)
			if err != nil {
				return err
			}

			log.Printf("Вкинули статью \"%s\" автора \"%s\" в чате %d", permalink, author, chatID)

			msg.Text = "Голосование за пост открыто\n" + getInstantViewLink(author, permalink)
			markup := getVoteMarkup(voteID, 0, 0)
			msg.ReplyMarkup = markup
			msg.DisableWebPagePreview = false
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
			return nil
		case state.Action == addKeyButtonText:
			login := update.Message.Text
			credential := models.Credential{
				UserID:   userID,
				UserName: login,
				Rating:   config.InitialUserRating,
			}
			if rating, err := credential.GetRating(database); err == nil {
				credential.Rating = rating
			}

			golos := client.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			if golos.LoginWithAuths(credential.UserName, config.PostingKey, []string{}, []string{config.Account}) {
				result, err := credential.Save(database)
				if err != nil {
					return err
				}
				if result {
					msg.Text = "Логин и приватный ключ успешно сохранён! " +
						"Присоединяйтесь к нашей группе для участия в курировании: " + config.GroupLink
				} else {
					log.Printf("Не сохранился приватный ключ: %#v", credential)
					msg.Text = "Не смог сохранить логин и приватный ключ :("
				}
			} else {
				msg.Text = "Мой публичный ключик ещё не добавлен в твой аккаунт."
			}
		default:
			if update.Message.Chat.Type != "private" {
				return nil
			}
			msg.Text = "Не понимаю"
		}
	} else if update.CallbackQuery != nil {
		arr := strings.Split(update.CallbackQuery.Data, "_")
		voteStringID, action := arr[0], arr[1]
		voteID, err := strconv.ParseInt(voteStringID, 10, 64)
		if err != nil {
			return err
		}

		if models.GetLastResponse(database).UserID == userID {
			config := tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Нельзя так часто голосовать",
			}
			bot.AnswerCallbackQuery(config)
			return nil
		}

		voteModel := models.GetVote(database, voteID)
		if voteModel.Completed {
			return nil
		}
		if voteModel.UserID == userID {
			config := tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Нельзя голосовать за свой же пост!",
			}
			bot.AnswerCallbackQuery(config)
			return nil
		}

		isGood := action == "good"
		response := models.Response{
			UserID: userID,
			VoteID: voteID,
			Result: isGood,
			Date:   time.Now(),
		}
		text := "Вы уже голосовали!"
		responseExists := response.Exists(database)
		if !responseExists {
			text = "Голос принят"
		}

		credential := models.Credential{UserID: userID}
		rating, err := credential.GetRating(database)
		if err != nil {
			return err
		}
		if rating <= config.RequiredVotes {
			text = "Слишком мало рейтинга для голосования, предлагайте посты"
			config := tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            text,
			}
			bot.AnswerCallbackQuery(config)
			return nil
		}

		callbackConfig := tgbotapi.CallbackConfig{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            text,
		}
		bot.AnswerCallbackQuery(callbackConfig)

		if !responseExists {
			_, err := response.Save(database)
			if err != nil {
				return err
			}
			voteModel := models.GetVote(database, voteID)
			err = verifyVotes(bot, voteModel, update, config, database)
			if err != nil {
				return err
			}
			// уменьшаем рейтинг голосовавшего при отрциательном голосовании
			if !response.Result {
				credential.DecrementRating(database, 1)
			}
		}
		return nil
	}

	_, err = state.Save(database)
	if err != nil {
		return err
	}
	log.Printf("%v", state)

	if msg.Text == "" {
		return errors.New("отсутствует текст сообщения")
	}
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func verifyVotes(bot *tgbotapi.BotAPI, voteModel models.Vote, update tgbotapi.Update, config config.Config, database *sql.DB) error {
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

	if positives+negatives >= config.RequiredVotes {
		if voteModel.Completed {
			return nil
		}
		voteModel.Completed = true
		_, err := voteModel.Save(database)
		if err != nil {
			return err
		}
		msg := tgbotapi.NewEditMessageText(chatID, messageID, "")
		if positives >= negatives {
			credential.IncrementRating(database, 1)
			successVotes := vote(voteModel, config, database)
			msg.Text = fmt.Sprintf("Проголосовала с силой %d%% c %d аккаунтов", voteModel.Percent, successVotes)
		} else {
			credential.DecrementRating(database, 2*config.RequiredVotes)
			rating, err := credential.GetRating(database)
			if err != nil {
				return err
			}
			msg.Text = "Пост отклонен, рейтинг предлагающего снижен"
			if rating < 0 {
				err = removeUser(bot, chatID, userID)
				if err != nil {
					log.Println(err)
					msg.Text = "Пост отклонен, предлагающий должен быть исключен"
				} else {
					msg.Text = "Пост отклонен, предлагающий исключен"
				}
			}
			// восстанавливаем рейтинг кураторам
			for _, response := range responses {
				// которые отклонили пост
				if false == response.Result {
					credential, err := models.GetCredentialByUserID(response.UserID, database)
					if err != nil {
						return err
					}
					err = credential.IncrementRating(database, 1)
					if err != nil {
						return err
					}
				}
			}
		}
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeUser(bot *tgbotapi.BotAPI, chatID int64, userID int) error {
	memberConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: 0,
	}
	_, err := bot.KickChatMember(memberConfig)
	return err
}

func vote(vote models.Vote, config config.Config, database *sql.DB) int {
	credentials, err := models.GetAllCredentials(database)
	if err != nil {
		log.Println("Не смогли извлечь ключи из базы")
		return 0
	}
	for _, credential := range credentials {
		client.Key_List[credential.UserName] = client.Keys{PKey: config.PostingKey}
	}
	log.Printf("Загружено %d аккаунтов", len(credentials))

	var errors []error
	var wg sync.WaitGroup
	wg.Add(len(credentials))
	for _, credential := range credentials {
		client.Key_List[credential.UserName] = client.Keys{PKey: config.PostingKey}
		go func(credential models.Credential) {
			defer wg.Done()
			weight := vote.Percent * 100
			golos := client.NewApi(config.Rpc, config.Chain)
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
	goodButton := tgbotapi.NewInlineKeyboardButtonData("👍Поддержать ("+strconv.Itoa(positives)+")", stringVoteID+"_good")
	badButton := tgbotapi.NewInlineKeyboardButtonData("👎Отклонить ("+strconv.Itoa(negatives)+")", stringVoteID+"_bad")
	row := []tgbotapi.InlineKeyboardButton{badButton, goodButton}
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

func getInstantViewLink(author string, permalink string) string {
	return "https://t.me/iv?url=https://goldvoice.club/" + "@" + author + "/" + permalink + "&rhash=70f46c6616076d"
}
