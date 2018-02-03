package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	golosClient "github.com/asuleymanov/golos-go/client"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/grokify/html-strip-tags-go"

	configuration "github.com/GolosTools/golos-vote-bot/config"
	"github.com/GolosTools/golos-vote-bot/db"
	"github.com/GolosTools/golos-vote-bot/helpers"
	"github.com/GolosTools/golos-vote-bot/models"
)

const (
	buttonAddKey        = "🐬Делегировать"
	buttonRemoveKey     = "🦀Остановить"
	buttonSetPowerLimit = "💪Настройка"
	buttonInformation   = "⚓️Информация"
)

var (
	config   configuration.Config
	database *sql.DB
	bot      *tgbotapi.BotAPI
)

func main() {
	configuration, err := helpers.GetConfig()
	if err != nil {
		log.Panic(err.Error())
	}
	config = configuration
	golosClient.Key_List[config.Account] = golosClient.Keys{
		PKey: config.PostingKey,
		AKey: config.ActiveKey}
	database, err = db.InitDB(config.DatabasePath)
	if err != nil {
		if err.Error() == "unable to open database file" {
			path, err := filepath.Abs(config.DatabasePath)
			if err != nil {
				log.Panic(err)
			}
			log.Panic(fmt.Sprintf("unable to open database at path: %s", path))
		}
		log.Panic(err)
	}
	defer database.Close()

	bot, err = tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = config.DebugMode
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go freshnessPolice()
	go checkAuthority()
	go queueProcessor()
	//go supportedPostsReporter()
	//go curationMotivator()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}
	for update := range updates {
		err := processMessage(update)
		if err != nil {
			log.Println(err)
		}
	}
}

func processMessage(update tgbotapi.Update) error {
	chatID, err := helpers.GetChatID(update)
	if err != nil {
		return err
	}
	userID, err := helpers.GetUserID(update)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatID, "")

	state, err := models.GetStateByUserID(userID, database)
	if err != nil {
		return err
	}

	if update.Message != nil {
		domainRegexp, err := helpers.GetDomainRegexp(config.Domains)
		if err != nil {
			return err
		}
		if update.Message.Chat.Type != "private" {
			return nil
		}
		switch {
		case update.Message.IsCommand():
			switch update.Message.Command() {
			case "start":
				username := "%username%"
				if len(update.Message.From.FirstName) > 0 {
					username = update.Message.From.FirstName
				}
				msg.Text = fmt.Sprintf("Привет, %s! \n\n"+
					"Я — бот для коллективного кураторства в [социальной блокчейн-сети \"Голос\"](https://golos.io).\n\n"+
					"Мой код полностью открыт и находится здесь: %s\n\n"+
					"Предлагаю начать с нажатия кнопки \""+buttonAddKey+"\", "+
					"после чего я дам ссылку на группу для предложения постов.\n\n"+
					"По любым вопросам пиши моему хозяину — %s",
					username, config.Repository, config.Developer)
				// save referral if exists
				if len(update.Message.CommandArguments()) > 0 {
					_, err := models.GetCredentialByUserID(userID, database)
					if err == sql.ErrNoRows {
						decodedString, err := base64.URLEncoding.DecodeString(update.Message.CommandArguments())
						if err == nil {
							referrer, err := models.GetCredentialByUserName(string(decodedString), database)
							if err == nil && referrer.Active == true {
								referral := models.Referral{UserID: userID,
									Referrer:  string(decodedString),
									UserName:  "",
									Completed: false}
								_, err = referral.Save(database)
								if err != nil {
									log.Println("не сохранили реферала: " + err.Error())
								}
							}
						} else {
							log.Printf("не смогли раскодировать строку %s", update.Message.CommandArguments())
						}
					}
				}
			}
			state.Action = update.Message.Command()
		case update.Message.Text == buttonAddKey:
			msg.Text = fmt.Sprintf("Добавь доверенный аккаунт *%s* в "+
				"[https://golostools.github.io/golos-vote-bot/](https://golostools.github.io/golos-vote-bot/) "+
				"(или через [форму от vik'a](https://golos.cf/multi/)), "+
				"а затем скажи мне свой логин на Голосе", config.Account)
			state.Action = buttonAddKey
		case update.Message.Text == buttonRemoveKey:
			msg.Text = fmt.Sprintf("Произошла ошибка, свяжись с разработчиком - %s", config.Developer)
			isActive := models.IsActiveCredential(userID, database)
			if isActive {
				credential, err := models.GetCredentialByUserID(userID, database)
				credential.Active = false
				credential.Curates = false
				result, err := credential.Save(database)
				if err != nil {
					log.Println(err.Error())
				}
				if true == result && err == nil {
					msg.Text = "Отлично, я больше не буду использовать твой аккаунт при курировании постов. " +
						"Дополнительно можешь удалить все сторонние ключи из своего аккаунта здесь: " +
						"https://golos.cf/multi/off.html"
				}
			} else {
				msg.Text = "Аккаунт не активирован"
			}
			state.Action = buttonRemoveKey
		case update.Message.Text == buttonSetPowerLimit:
			if false == models.IsActiveCredential(userID, database) {
				msg.Text = "Сначала делегируй мне права кнопкой " + buttonAddKey
				break
			}
			msg.Text = "Введи значение делегируемой силы Голоса от 1 до 100%"
			state.Action = buttonSetPowerLimit
		case update.Message.Text == buttonInformation:
			if false == models.IsActiveCredential(userID, database) {
				msg.Text = "У меня пока нет информации для тебя"
				break
			}
			credential, err := models.GetCredentialByUserID(userID, database)
			if err != nil {
				return err
			}
			encodedUserName := base64.URLEncoding.EncodeToString([]byte(credential.UserName))
			referralLink := "https://t.me/" + config.TelegramBotName + "?start=" + encodedUserName
			msg.Text = fmt.Sprintf("Аккаунт: *%s*, делегированная сила: *%d%%*\n"+
				"Реферальная ссылка: [%s](%s)\n"+
				"(дает обоим по %.3f Силы Голоса, "+
				"у приглашаемого должно быть как минимум %d постов "+
				"и он не должен взаимодействовать с Голосовалочкой до приглашения)",
				credential.UserName, credential.Power, referralLink, referralLink, config.ReferralFee, config.ReferralMinimumPostCount)
			var button tgbotapi.InlineKeyboardButton
			if models.IsActiveCurator(userID, database) {
				button = tgbotapi.NewInlineKeyboardButtonData("Прекратить кураторство", "curating_stop")
			} else {
				button = tgbotapi.NewInlineKeyboardButtonData("Стать куратором", "curating_start")
			}
			keyboard := []tgbotapi.InlineKeyboardButton{button}
			markup := tgbotapi.NewInlineKeyboardMarkup(keyboard)
			msg.ReplyMarkup = markup
			state.Action = buttonInformation
		case domainRegexp.MatchString(update.Message.Text):
			msg.ReplyToMessageID = update.Message.MessageID

			matched := domainRegexp.FindStringSubmatch(update.Message.Text)
			author, permalink := matched[1], matched[2]

			golos := golosClient.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			post, err := golos.Rpc.Database.GetContent(author, permalink)
			if err != nil {
				return err
			}
			// check post exists in blockchain
			if post.Author != author || post.Permlink != permalink {
				return nil
			}

			lastVote := models.GetLastVoteForUserID(userID, database)
			userInterval, _ := models.ComputeIntervalForUser(userID, 10, config.PostingInterval, database)
			if time.Since(lastVote.Date) < userInterval && !config.DebugMode {
				msg.Text = "Прошло слишком мало времени после твоего последнего поста. Наберись терпения!"
				break
			}

			if config.Censorship {
				tags := post.JsonMetadata.Tags
				includesBannedTag := false
				for _, bannedTag := range config.BannedTags {
					for _, postTag := range tags {
						if postTag == bannedTag {
							includesBannedTag = true
							msg.Text = "Нельзя предлагать посты с тегом " + postTag
						}
					}

				}
				if includesBannedTag {
					break
				}
			}

			isActive := models.IsActiveCredential(userID, database)
			if !isActive {
				msg.Text = "Предлагать посты для голосования могут только голосующие пользователи. Жулик не воруй!"
				break
			}

			if post.Mode != "first_payout" {
				msg.Text = "Выплата за пост уже была произведена! Есть что-нибудь посвежее?"
				break
			}

			if post.MaxAcceptedPayout == "0.000 GBG" {
				msg.Text = "Мне не интересно голосовать за пост с отключенными выплатами"
				break
			}

			if models.GetOpenedVotesCount(database) >= config.MaximumOpenedVotes {
				msg.Text = "Слишком много уже открытых голосований. " +
					"Подожди, пока другой голос получит голоса или полиция свежести избавится от протухших постов."
				break
			}

			if helpers.IsVoxPopuli(author) && config.IgnoreVP {
				msg.Text = "Сообщества vox-populi могут сами себя поддержать"
				break
			}

			if len(post.Body) < config.MinimumPostLength {
				msg.Text = "Слишком мало текста, не скупись на буквы!"
				break
			}

			percent := 100

			voteModel := models.Vote{
				UserID:    userID,
				Author:    author,
				Permalink: permalink,
				Percent:   percent,
				Completed: false,
				Rejected:  false,
				Addled:    false,
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

			msg.Text = "Пост выставлен на голосование."

			if checkUniqueness(post.Body, voteModel) {
				go newPost(voteID, author, permalink, chatID)
			}

			return nil
		case state.Action == buttonAddKey:
			login := strings.ToLower(update.Message.Text)
			login = strings.Trim(login, "@")
			credential := models.Credential{
				UserID:   userID,
				ChatID:   chatID,
				UserName: login,
				Power:    100,
				Active:   true,
				Curates:  false,
			}

			golos := golosClient.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			accounts, err := golos.Rpc.Database.GetAccounts([]string{login})
			if err != nil {
				return err
			} else if len(accounts) == 1 {
				hasPostingAuth := false
				for _, auth := range accounts[0].Posting.AccountAuths {
					if auth.([]interface{})[0] == config.Account {
						hasPostingAuth = true
						break
					}
				}
				if hasPostingAuth {
					// send referral fee
					referral, err := models.GetReferralByUserID(userID, database)
					if err == nil && !referral.Completed {
						if err = referral.SetCompleted(database); err == nil {
							referral.UserName = credential.UserName
							referral.Save(database)
							_, err = models.GetCredentialByUserName(credential.UserName, database)
							if err == sql.ErrNoRows {
								go sendReferralFee(referral.Referrer, credential.UserName)
							}
						}
					}

					_, err = credential.Save(database)
					if err != nil {
						return err
					}
					msg.Text = "Поздравляю, теперь ты почти полноправный участник! " +
						"Чтобы вообще все плюшки заиметь, можешь стать еще и куратором. " +
						"Присоединяйся к нашей группе, там бывает весело: " + config.GroupLink
					state.Action = "successAuth"
				} else {
					msg.Text = fmt.Sprintf("Доступ у этого аккаунта для меня отсутствует. "+
						"Добавить его можно в https://golos.cf/multi/ для аккаунта *%s*", config.Account)
				}
			} else {
				msg.Text = fmt.Sprintf("Что-то пошло не так. Попробуй повторить позже "+
					"или свяжись с разработчиком: %s", config.Developer)
				log.Printf("Введён некорректный логин: %s", update.Message.Text)
			}
		case state.Action == buttonSetPowerLimit:
			re := regexp.MustCompile("[0-9]+")
			valueString := re.FindString(update.Message.Text)
			value, err := strconv.Atoi(valueString)
			if err != nil {
				msg.Text = "Не поняла. Введи значение делегируемой силы Голоса от 1 до 100%"
				break
			}
			if value >= 1 && value <= 100 {
				if false == models.IsActiveCredential(userID, database) {
					msg.Text = "Сначала делегируй мне права кнопкой " + buttonAddKey
					break
				}

				credential, err := models.GetCredentialByUserID(userID, database)
				if err != nil {
					return err
				}

				golos := golosClient.NewApi(config.Rpc, config.Chain)
				defer golos.Rpc.Close()

				accounts, err := golos.Rpc.Database.GetAccounts([]string{credential.UserName})
				if err != nil {
					return err
				}

				voteWeightThreshold := 1.0 * 1000.0 * 1000.0
				vestingSharesPreparedString := strings.Split(accounts[0].VestingShares, " ")[0]
				vestingShares, err := strconv.ParseFloat(vestingSharesPreparedString, 64)
				if err != nil {
					return err
				}
				if vestingShares > voteWeightThreshold {
					err = credential.UpdatePower(value, database)
					if err != nil {
						return err
					}
					msg.Text = fmt.Sprintf("Предоставленная мне в распоряжение сила Голоса "+
						"для аккаунта *%s* теперь равна *%d%%*", credential.UserName, value)
				} else {
					msg.Text = "У тебя пока слишком маленькая Сила Голоса для этого"
				}
				state.Action = "updatedPower"
			}
		default:
			if update.Message.Chat.Type != "private" {
				return nil
			}
			msg.ReplyToMessageID = update.Message.MessageID
			msg.Text = "Не понимаю"
		}
		if msg.ReplyMarkup == nil && update.Message.Chat.Type == "private" {
			firstButton := tgbotapi.NewKeyboardButton(buttonAddKey)
			secondButton := tgbotapi.NewKeyboardButton(buttonRemoveKey)
			firstButtonRow := []tgbotapi.KeyboardButton{firstButton, secondButton}

			thirdButton := tgbotapi.NewKeyboardButton(buttonSetPowerLimit)
			fourthButton := tgbotapi.NewKeyboardButton(buttonInformation)
			secondButtonRow := []tgbotapi.KeyboardButton{thirdButton, fourthButton}

			keyboard := tgbotapi.NewReplyKeyboard(firstButtonRow, secondButtonRow)
			msg.ReplyMarkup = keyboard
		}
	} else if update.CallbackQuery != nil {
		arr := strings.Split(update.CallbackQuery.Data, "_")
		voteStringID, action := arr[0], arr[1]
		if voteStringID == "curating" {
			switch action {
			case "start":
				if models.IsActiveCurator(userID, database) {
					msg.Text = "Ты уже являешься куратором"
					bot.Send(msg)
					return nil
				}
				credential, err := models.GetCredentialByUserID(userID, database)
				if err == nil && credential.ChatID == 0 {
					credential.ChatID = chatID
					credential.Save(database)
				}
				msg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "")
				msg.Text = config.CurationRules
				approveButton := tgbotapi.NewInlineKeyboardButtonData("🐬‍️Я справлюсь", "curating_approve")
				declineButton := tgbotapi.NewInlineKeyboardButtonData("🐡‍Слишком сложно", "curating_decline")
				keyboard := []tgbotapi.InlineKeyboardButton{approveButton, declineButton}
				markup := tgbotapi.NewInlineKeyboardMarkup(keyboard)
				msg.ReplyMarkup = &markup
				bot.Send(msg)
			case "approve":
				err = models.ActivateCurator(userID, database)
				if err != nil {
					return err
				}
				msg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "")
				msg.Text = "Отлично, теперь ты будешь участвовать в курировании постов. " +
					"Скоро я начну присылать тебе ссылки, подожди немного"
				bot.Send(msg)
			case "decline":
				msg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "")
				msg.Text = "Хороший выбор. Курирование чужих постов — сложный и неблагодарный процесс. " +
					"Лучше пиши свои посты и скидывай мне ссылки на них, а кураторы пусть делают свою работу!"
				bot.Send(msg)
			case "stop":
				msg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "")
				if models.IsActiveCurator(userID, database) {
					err = models.DeactivateCurator(userID, database)
					if err != nil {
						return nil
					}
					msg.Text = "Бремя кураторства покинуло тебя. Когда вдоволь насладишься свободой — возвращайся!"
				} else {
					msg.Text = "То, что мертво — умереть не может. Так и ты — нельзя отказаться от курирования, не будучи куратором"
				}
				bot.Send(msg)
			default:
				return errors.New("неподдерживаемое действие: " + action)
			}
		} else {
			voteID, err := strconv.ParseInt(voteStringID, 10, 64)
			if err != nil {
				return err
			}
			if !models.IsActiveCurator(userID, database) {
				config := tgbotapi.CallbackConfig{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Чекни свои привелегии. Ты не куратор!",
				}
				bot.AnswerCallbackQuery(config)
				return nil
			}

			voteModel := models.GetVote(database, voteID)
			if voteModel.Completed {
				return nil
			}

			isGood := action == "good"
			response := models.Response{
				UserID: userID,
				VoteID: voteID,
				Result: isGood,
				Date:   time.Now(),
			}
			text := "И да настигнет Админская кара всех тех, кто пытается злоупотреблять своей властью и голосовать несколько раз! Админь"
			responseExists := response.Exists(database)
			if !responseExists {
				text = "Голос принят"
				messageID, err := helpers.GetMessageID(update)
				if err != nil {
					return err
				}
				msg := tgbotapi.NewEditMessageText(chatID, messageID, "")
				msg.Text = text
				_, err = bot.Send(msg)
				if err != nil {
					log.Println(err.Error())
				}
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
			}
		}
		return nil
	}

	_, err = state.Save(database)
	if err != nil {
		return err
	}

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

// https://text.ru/api-check/manual
func checkUniqueness(text string, voteModel models.Vote) bool {
	token := config.TextRuToken
	if len(config.TextRuToken) == 0 {
		return false
	}

	text = strip.StripTags(text)

	if len(text) < config.MinimumPostLength {
		return false
	}

	cut := func(text string, to int) string {
		runes := []rune(text)
		if len(runes) > to {
			return string(runes[:to])
		}
		return text
	}
	maxSymbolCount := 2000
	text = cut(text, maxSymbolCount)

	httpClient := http.Client{}
	form := url.Values{}
	form.Add("text", text)
	form.Add("userkey", token)
	domainList := strings.Join(config.Domains, ",")
	form.Add("exceptdomain", domainList)
	form.Add("visible", "vis_on")
	req, err := http.NewRequest("POST", "http://api.text.ru/post", strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err.Error())
		return false
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		log.Println("статус не 200")
		return false
	}
	type Uid struct {
		TextUid string `json:"text_uid"`
	}
	var uid Uid
	jsonParser := json.NewDecoder(resp.Body)
	jsonParser.Decode(&uid)
	if len(uid.TextUid) == 0 {
		log.Println("Не распарсили text_uid")
		return false
	}
	step := 0
	for step < 50 {
		step += 1
		time.Sleep(time.Second * 15)
		log.Printf("step %d", step)
		client := http.Client{}
		form := url.Values{}
		form.Add("uid", uid.TextUid)
		form.Add("userkey", token)
		//form.Add("jsonvisible", "detail")
		req, err := http.NewRequest("POST", "http://api.text.ru/post", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		type Result struct {
			TextUnique string `json:"text_unique"`
			ResultJson string `json:"result_json"`
		}
		var result Result
		jsonParser := json.NewDecoder(resp.Body)
		jsonParser.Decode(&result)
		if len(result.TextUnique) == 0 {
			continue
		}
		textUnique, err := strconv.ParseFloat(result.TextUnique, 32)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		log.Println(textUnique)
		if textUnique < 20 {
			voteModel.Completed = true
			_, err := voteModel.Save(database)
			if err != nil {
				log.Println(err.Error())
				return false
			}
			return false
		} else {
			random := func(min, max int) int {
				rand.Seed(time.Now().Unix())
				return rand.Intn(max-min) + min
			}
			imageNumber := random(1, 18)
			report := fmt.Sprintf("[![Уникальность проверена через TEXT.RU](https://text.ru/image/get/%s/%d)](https://text.ru/antiplagiat/%s)",
				uid.TextUid, imageNumber, uid.TextUid)
			err = helpers.SendComment(voteModel.Author, voteModel.Permalink, report, config)
			if err != nil {
				log.Println(err.Error())
			}
			return true
		}
		// если дошли сюда, то выходим из цикла
		break
	}
	return false
}

func sendReferralFee(referrer string, referral string) {
	if referrer == referral {
		log.Printf("Пригласивший и приглашенный %s совпадают", referral)
		return
	}
	golos := golosClient.NewApi(config.Rpc, config.Chain)
	defer golos.Rpc.Close()
	accounts, err := golos.Rpc.Database.GetAccounts([]string{referral})
	if err != nil {
		log.Println("Не получили аккаунт " + referral)
		return
	}
	if accounts[0].PostCount.Int64() < int64(config.ReferralMinimumPostCount) {
		log.Printf("За новичка %s награды не будет, слишком мало постов", referral)
		return
	}
	amount := fmt.Sprintf("%.3f GOLOS", config.ReferralFee)
	err = golos.TransferToVesting(config.Account, referrer, amount)
	err2 := golos.TransferToVesting(config.Account, referral, amount)
	if err != nil {
		log.Println(fmt.Sprintf("Не отправили силу голоса %s \nаккаунту %s", err.Error(), referrer))
	}
	if err2 != nil {
		log.Println(fmt.Sprintf("Не отправили силу голоса %s \nаккаунту %s", err.Error(), referral))
	}
	if err != nil || err2 != nil {
		return
	}
	markdownLink := func(account string) string {
		return fmt.Sprintf("[@%s](https://golos.io/@%s/transfers)", account, account)
	}
	referrerLink := markdownLink(referrer)
	referralLink := markdownLink(referral)
	text := fmt.Sprintf("Пригласивший %s и приглашённый %s получают по %.3f Силы Голоса в рамках партнёрской программы",
		referrerLink, referralLink, config.ReferralFee)
	msg := tgbotapi.NewMessage(config.GroupID, text)
	msg.ParseMode = "Markdown"
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Не отправили сообщение: " + err.Error())
	}
}

func checkAuthority() {
	for {
		credentials, err := models.GetAllActiveCredentials(database)
		log.Printf("Загружено %d аккаунтов для проверки", len(credentials))
		if err != nil {
			log.Println(err.Error())
		}
		golos := golosClient.NewApi(config.Rpc, config.Chain)
		for _, credential := range credentials {
			if !golos.Verify_Delegate_Posting_Key_Sign(credential.UserName, config.Account) {
				log.Printf("Пользователь %s отключён", credential.UserName)
				credential.Active = false
				_, err = credential.Save(database)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
		golos.Rpc.Close()
		time.Sleep(1 * time.Hour)
	}
}

func newPost(voteID int64, author string, permalink string, chatID int64) {
	curatorChatIDs, err := models.GetAllActiveCurstorsChatID(database)
	if err != nil {
		log.Println(err.Error())
		return
	}
	curateText := "Новый пост - новая оценка. Курируй, куратор\n" + helpers.GetInstantViewLink(author, permalink)
	for _, curatorChatID := range curatorChatIDs {
		if curatorChatID == chatID {
			continue
		}
		msg := tgbotapi.NewMessage(curatorChatID, curateText)
		markup := helpers.GetVoteMarkup(voteID)
		msg.ReplyMarkup = markup
		msg.DisableWebPagePreview = false

		_, err := bot.Send(msg)
		if err != nil {
			log.Println(fmt.Sprintf("Не смогли отправить сообщение куратору %n", curatorChatID))
		}
	}
}

func queueProcessor() {
	for {
		// TODO: вынести минуты в настройки
		time.Sleep(36 * time.Minute)
		log.Println("Начинаю голосование за лучший пост")
		votes, err := models.GetAllOpenedVotes(database)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if len(votes) == 0 {
			log.Println("Нет открытых голосований")
			continue
		}
		maxDiff := 0
		mostLikedPost := votes[0]
		for _, vote := range votes {
			var positives, negatives int
			positives, negatives = models.GetNumResponsesVoteID(vote.VoteID, database)
			diff := positives - negatives
			if diff > maxDiff {
				maxDiff = diff
				mostLikedPost = vote
			}
		}
		log.Printf("Лучший пост определен: %s/%s", mostLikedPost.Author, mostLikedPost.Permalink)
		successVotesCount, err := helpers.Vote(mostLikedPost, database, config)
		text := fmt.Sprintf("Успешно проголосовала c %d аккаунтов за пост\n%s",
			successVotesCount,
			helpers.GetInstantViewLink(mostLikedPost.Author, mostLikedPost.Permalink))
		if err != nil {
			log.Println(err.Error())
			text = fmt.Sprintf("В процессе голосования произошла ошибка, свяжитесь с разработчиком - %s\n%s",
				config.Developer,
				helpers.GetInstantViewLink(mostLikedPost.Author, mostLikedPost.Permalink))
		}
		msg := tgbotapi.NewMessage(config.GroupID, text)
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func freshnessPolice() {
	golos := golosClient.NewApi(config.Rpc, config.Chain)
	votes, err := models.GetAllOpenedVotes(database)
	log.Printf("Загружено %d постов для проверки", len(votes))
	if err != nil {
		log.Panic(err.Error())
	}
	for _, vote := range votes {
		post, err := golos.Rpc.Database.GetContent(vote.Author, vote.Permalink)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if post.Mode != "first_payout" {
			vote.Completed = true
			vote.Addled = true
			vote.Save(database)
			go excuseUs(vote)
		}
	}
	golos.Rpc.Close()
	time.Sleep(1 * time.Hour)
	freshnessPolice()
}

func excuseUs(vote models.Vote) {
	positives, negatives := models.GetNumResponsesVoteID(vote.VoteID, database)
	var msg tgbotapi.MessageConfig
	if positives >= negatives {
		text := fmt.Sprintf("Прости, %s, твой пост (%s/%s) так и не дождался своих голосов. В следующий раз напиши что-нибудь "+
			"получше и кураторы обязательно это оценят", vote.Author, vote.Author, vote.Permalink)
		msg = tgbotapi.NewMessage(config.GroupID, text)
	} else {
		vote.Rejected = true
		vote.Save(database)
		text := fmt.Sprintf("Пoст %d/%d был отклонен кураторами", vote.Author, vote.Permalink)
		msg = tgbotapi.NewMessage(config.GroupID, text)
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func supportedPostsReporter() {
	time.Sleep(models.WannaSleepOneDay(12, 0)) // Спать до 12:00 следующего дня
	for {
		//supportedPosts, err:= models.GetTrulyCompletedVotesSince(models.GetLastReportDate(database), database)
		//if err != nil {
		//	log.Println(err)
		//} else {
		//Я понятия не имею, как постить пост
		//err := golos.Post(config.Account, title, body, permlink, "", post_image string, config.ReportTags, v *PC_Vote, o *PC_Options)
		//if err != nil {
		//	log.Println(err)
		//} else {
		//	models.NewReportPosted(database)
		//}

		//}
		time.Sleep(24 * time.Hour)
	}
}

func curationMotivator() {
	time.Sleep(models.WannaSleepTill(0, 20, 0)) // Спать до 20:00 ближайшего воскресенья
	for {
		lastRewardDate := models.GetLastRewardDate(database)
		allResponses := models.GetNumResponsesForMotivation(lastRewardDate, database)
		var needResponsesToBeRewarded int

		golos := golosClient.NewApi(config.Rpc, config.Chain)
		defer golos.Rpc.Close()
		accounts, err := golos.Rpc.Database.GetAccounts([]string{config.Account})
		if err != nil {
			log.Println(err)
		} else {
			gold, _ := strconv.Atoi(strings.Replace(strings.Replace(accounts[0].SbdBalance, ".", "", 1), " GBG", "", 1))
			if gold < allResponses {
				needResponsesToBeRewarded = allResponses / gold
			} else {
				needResponsesToBeRewarded = 1
			}
			curatorIDs, err := models.GetUserIDsForMotivation(lastRewardDate, database)
			if err != nil {
				log.Println(err)
			} else {
				for _, userID := range curatorIDs {
					credential, err := models.GetCredentialByUserID(userID, database)
					if !credential.Active || err != nil {
						continue
					}
					curatorResponses := models.GetNumResponsesForMotivationForUserID(userID, lastRewardDate, database)
					goldForCurator := curatorResponses / needResponsesToBeRewarded
					amount := fmt.Sprintf("%d.%.3d GBG", goldForCurator/1000, goldForCurator%1000)
					err = golos.Transfer(config.Account, credential.UserName, "Вознаграждение для кураторов", amount)
				}
			}
		}
		time.Sleep(7 * 24 * time.Hour)
	}
}
