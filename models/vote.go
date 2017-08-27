package models

import (
	"database/sql"
)

type Vote struct {
	UserID    int
	Voter     string
	Author    string
	Permalink string
	Percent   int
}

func (vote Vote) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT INTO votes(" +
		"user_id," +
		"author," +
		"permalink," +
		"percent) " +
		"values(?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(vote.UserID, vote.Author, vote.Permalink, vote.Percent)
	if err != nil {
		return false, err
	}
	return true, nil
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
