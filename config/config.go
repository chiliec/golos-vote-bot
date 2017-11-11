package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	TelegramToken          string   `json:"telegram_token"`
	Account                string   `json:"account"`
	PostingKey             string   `json:"posting_key"`
	ActiveKey              string   `json:"active_key"`
	RequiredVotes          int      `json:"required_votes"`
	InitialUserRating      int      `json:"initial_user_rating"`
	MaximumOpenedVotes     int      `json:"maximum_opened_votes"`
	MaximumUserVotesPerDay int      `json:"maximum_user_votes_per_day"`
	Developer              string   `json:"developer"`
	GroupID                int64    `json:"group_id"`
	GroupLink              string   `json:"group_link"`
	DatabasePath           string   `json:"database_path"`
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
