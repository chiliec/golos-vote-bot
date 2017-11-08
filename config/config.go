package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	TelegramToken          string   `json:"telegramToken"`
	RequiredVotes          int      `json:"requiredVotes"`
	InitialUserRating      int      `json:"initialUserRating"`
	MaximumOpenedVotes     int      `json:"maximumOpenedVotes"`
	MaximumUserVotesPerDay int      `json:"maximumUserVotesPerDay"`
	Developer              string   `json:"developer"`
	GroupID                int64    `json:"groupID"`
	GroupLink              string   `json:"groupLink"`
	Chain                  string   `json:"chain"`
	Rpc                    []string `json:"rpc"`
}

func LoadConfiguration(file string, config *Config) error {
	configFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return nil
}
