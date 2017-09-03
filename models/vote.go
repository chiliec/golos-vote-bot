package models

import (
	"database/sql"
)

type Vote struct {
	UserID    int
	Author    string
	Permalink string
	Percent   int
}

func (vote Vote) Save(db *sql.DB) (int64, error) {
	prepare, err := db.Prepare("INSERT INTO votes(" +
		"user_id," +
		"author," +
		"permalink," +
		"percent) " +
		"values(?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	result, err := prepare.Exec(vote.UserID, vote.Author, vote.Permalink, vote.Percent)
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
