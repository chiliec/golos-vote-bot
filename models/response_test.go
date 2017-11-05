package models

import (
	"testing"
	"time"

	"github.com/GolosTools/golos-vote-bot/db"
)

func TestDbResponses(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Failed()
	}
	response := Response{
		UserID: 1,
		VoteID: 1,
		Result: true,
		Date:   time.Now(),
	}
	response.Save(database)
	responsesFromDB, err := GetAllResponsesForVoteID(response.VoteID, database)
	if response.Date.Unix() != responsesFromDB[0].Date.Unix() {
		t.Error("Даты не совпадают!")
	}
	// ссылочный тип
	response.Date = responsesFromDB[0].Date
	if response != responsesFromDB[0] {
		t.Errorf("\n%#v\n%#v\nНе равны!", response, responsesFromDB[0])
	}
}
