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

func GetLastResponse(db *sql.DB) (response Response) {
	row := db.QueryRow("SELECT user_id, vote_id, result, date FROM responses ORDER BY ID DESC LIMIT 1")
	row.Scan(&response.UserID, &response.VoteID, &response.Result, &response.Date)
	return response
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
