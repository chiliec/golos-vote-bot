package config

import (
	"reflect"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	var config Config
	err := LoadConfiguration("../config.json", &config)
	if err != nil {
		t.Error(err)
	}
	defaultConfig := Config{
		DebugMode:                false,
		TelegramToken:            "write-your-telegram-token-here",
		TelegramBotName:          "golosovalochka_bot",
		Account:                  "golosovalochka",
		PostingKey:               "5...",
		ActiveKey:                "5...",
		TextRuToken:              "",
		ReferralFee:              5.0,
		ReferralMinimumPostCount: 30,
		MaximumOpenedVotes:       3,
		PostingInterval:          480,
		MinimumPostLength:        1000,
		Developer:                "@babin",
		GroupID:                  -1001143551951,
		GroupLink:                "https://t.me/joinchat/AlKeQUQpN8-9oShtaTcY7Q",
		DatabasePath:             "./db/database.db",
		Domains:                  []string{"golos.io", "golos.blog", "goldvoice.club", "golosd.com", "golosdb.com", "mapala.net", "newbie.goloses.ru", "cpeda.space"},
		Chain:                    "golos",
		Rpc:                      []string{"wss://ws.golos.io", "wss://api.golos.cf"},
		Repository:               "https://github.com/GolosTools/golos-vote-bot",
		IgnoreVP:                 true,
		BannedTags:               []string{"test", "test1"},
		Censorship:               false,
		ReportTags:               []string{"тест", "тест1"},
		CurationRules:            "Правила курирования. Здесь нужно написать описание правил курирования!",
	}
	if !reflect.DeepEqual(defaultConfig, config) {
		t.Error("Конфиги не совпадают")
	}
}
