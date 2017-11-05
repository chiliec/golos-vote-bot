package models

import (
	"testing"
	"time"

	"github.com/GolosTools/golos-vote-bot/db"
)

func TestDbVotes(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Failed()
	}
	vote := Vote{
		VoteID:    1,
		UserID:    2,
		Author:    "ExampleAuthor",
		Permalink: "/example/permalink",
		Percent:   100,
		Completed: false,
		Date:      time.Now(),
	}
	vote.Save(database)
	voteFromDb := GetVote(database, vote.VoteID)
	if vote.Date.Unix() != voteFromDb.Date.Unix() {
		t.Error("Даты не совпадают!")
	}
	// ссылочный тип
	vote.Date = voteFromDb.Date
	if vote != voteFromDb {
		t.Errorf("\n%#v\n%#v\nНе равны!", vote, voteFromDb)
	}
}
