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
		Rejected:  false,
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

func TestGetLastVotes(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	firstVote := Vote{
		VoteID:    1,
		UserID:    1,
		Author:    "ExampleAuthor",
		Permalink: "/example/permalink",
		Percent:   100,
		Completed: false,
		Rejected:  false,
		Date:      time.Now(),
	}
	firstVote.Save(database)
	secondVote := Vote{
		VoteID:    2,
		UserID:    1,
		Author:    "ExampleAuthor2",
		Permalink: "/example/permalink2",
		Percent:   100,
		Completed: true,
		Rejected:  false,
		Date:      time.Now(),
	}
	secondVote.Save(database)

	lastVote := GetLastVoteForUserID(1, database)

	if secondVote.Date.Unix() != lastVote.Date.Unix() {
		t.Error("Даты не совпадают!")
	}

	// ссылочный тип
	firstVote.Date = lastVote.Date
	secondVote.Date = lastVote.Date

	if firstVote == lastVote {
		t.Errorf("\n%#v\n%#v\nРавны, а не должны быть!", firstVote, lastVote)
	}
	if secondVote != lastVote {
		t.Errorf("\n%#v\n%#v\nНе равны!", secondVote, lastVote)
	}
}
