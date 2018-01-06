package helpers

import (
	"errors"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetVoteMarkup(voteID int64, positives int, negatives int) tgbotapi.InlineKeyboardMarkup {
	stringVoteID := strconv.FormatInt(voteID, 10)
	goodButton := tgbotapi.NewInlineKeyboardButtonData("👍Лайк ("+strconv.Itoa(positives)+")", stringVoteID+"_good")
	badButton := tgbotapi.NewInlineKeyboardButtonData("👎Дизлайк ("+strconv.Itoa(negatives)+")", stringVoteID+"_bad")
	row := []tgbotapi.InlineKeyboardButton{badButton, goodButton}
	markup := tgbotapi.InlineKeyboardMarkup{}
	markup.InlineKeyboard = append(markup.InlineKeyboard, row)
	return markup
}

func GetChatID(update tgbotapi.Update) (int64, error) {
	if update.Message != nil {
		return update.Message.Chat.ID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID, nil
	} else {
		return 0, errors.New("не получили ID чата")
	}
}

func GetUserID(update tgbotapi.Update) (int, error) {
	if update.Message != nil {
		return update.Message.From.ID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID, nil
	} else {
		return 0, errors.New("не получили ID пользователя")
	}
}

func GetMessageID(update tgbotapi.Update) (int, error) {
	if update.Message != nil {
		return update.Message.MessageID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.MessageID, nil
	} else {
		return 0, errors.New("не получили ID сообщения")
	}
}

func GetInstantViewLink(author string, permalink string) string {
	return "https://t.me/iv?url=https://goldvoice.club/" + "@" + author + "/" + permalink + "&rhash=70f46c6616076d"
}
