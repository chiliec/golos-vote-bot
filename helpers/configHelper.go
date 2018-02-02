package helpers

import (
	"errors"
	"os"

	configuration "github.com/GolosTools/golos-vote-bot/config"
)

func GetConfig() (config configuration.Config, err error) {
	err = configuration.LoadConfiguration("./config.json", &config)
	if err != nil {
		return config, err
	}
	err = configuration.LoadConfiguration("./config.local.json", &config)
	if err != nil && !os.IsNotExist(err) {
		return config, err
	}
	if config.TelegramToken == "write-your-telegram-token-here" {
		return config, errors.New("токен для телеграма не введён")
	}
	return config, err
}
