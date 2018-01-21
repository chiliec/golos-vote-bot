package models

import (
	"database/sql"
	"time"
)

type Response struct {
	UserID int
	VoteID int64
	Result bool
	Date   time.Time
}

func (response Response) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO responses(" +
		"user_id," +
		"vote_id," +
		"result," +
		"date) " +
		"values(?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(response.UserID, response.VoteID, response.Result, response.Date)
	return err != nil, err
}

func (response Response) Exists(db *sql.DB) bool {
	row := db.QueryRow("SELECT id FROM responses "+
		"WHERE user_id = ? AND vote_id = ?", response.UserID, response.VoteID)
	var id *int
	row.Scan(&id)
	return id != nil
}

func GetAllResponsesForVoteID(voteID int64, db *sql.DB) (responses []Response, err error) {
	rows, err := db.Query("SELECT user_id, vote_id, result, date FROM responses WHERE vote_id = ?", voteID)
	if err != nil {
		return responses, err
	}
	defer rows.Close()
	for rows.Next() {
		var response Response
		rows.Scan(&response.UserID, &response.VoteID, &response.Result, &response.Date)
		responses = append(responses, response)
	}
	return responses, nil
}

func GetNumResponsesVoteID(voteID int64, db *sql.DB) (int, int) {
	var pos int
	var neg int
	row := db.QueryRow("SELECT COUNT(*) FROM responses WHERE vote_id = ? AND result = 1", voteID)
	row.Scan(&pos)
	row = db.QueryRow("SELECT COUNT(*) FROM responses WHERE vote_id = ? AND result = 0", voteID)
	row.Scan(&neg)
	return pos, neg
}

func GetNumResponsesForMotivation(date time.Time, db *sql.DB) (num int) {
	row := db.QueryRow("SELECT COUNT(*) FROM responses WHERE date > ?", date)
	row.Scan(&num)
	return num
}

func GetUserIDsForMotivation(date time.Time, db *sql.DB) (userIDs []int, err error) {
	rows, err := db.Query("SELECT distinct user_id FROM responses WHERE date > ?", date)
	if err != nil {
		return userIDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID int
		rows.Scan(&userID)
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func GetNumResponsesForMotivationForUserID(userID int, date time.Time, db *sql.DB) (num int) {
	row := db.QueryRow("SELECT COUNT(*) FROM responses WHERE date > ? AND user_id = ?", date, userID)
	row.Scan(&num)
	return num
}
