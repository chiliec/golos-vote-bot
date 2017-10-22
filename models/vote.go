package models

import (
	"database/sql"
)

type Vote struct {
	VoteID    int64
	UserID    int
	Author    string
	Permalink string
	Percent   int
	Completed bool
}

func GetVote(db *sql.DB, voteID int64) (vote Vote) {
	row := db.QueryRow("SELECT id, user_id, author, permalink, percent, completed FROM votes WHERE id = ?", voteID)
	row.Scan(&vote.VoteID, &vote.UserID, &vote.Author, &vote.Permalink, &vote.Percent, &vote.Completed)
	return vote
}

func (vote Vote) Save(db *sql.DB) (int64, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO votes(" +
		"user_id," +
		"author," +
		"permalink," +
		"percent," +
		"completed) " +
		"values(?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	result, err := prepare.Exec(vote.UserID, vote.Author, vote.Permalink, vote.Percent, vote.Completed)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (vote Vote) Exists(db *sql.DB) bool {
	row := db.QueryRow("SELECT user_id FROM votes WHERE author = ? AND permalink = ?", vote.Author, vote.Permalink)
	var userID *int
	row.Scan(&userID)
	if userID != nil {
		return true
	}
	return false
}
