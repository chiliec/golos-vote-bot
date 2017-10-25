package models

import (
	"database/sql"
)

type Response struct {
	UserID int
	VoteID int64
	Result bool
}

func (response Response) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO responses(" +
		"user_id," +
		"vote_id," +
		"result) " +
		"values(?, ?, ?)")
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(response.UserID, response.VoteID, response.Result)
	return err != nil, err
}

func (response Response) Exists(db *sql.DB) bool {
	row := db.QueryRow("SELECT id FROM responses WHERE user_id = ? AND vote_id = ?",
		response.UserID, response.VoteID)
	var id *int
	row.Scan(&id)
	return id != nil
}

func GetLastResponse(db *sql.DB) (response Response) {
	row := db.QueryRow("SELECT user_id, vote_id, result FROM responses ORDER BY ID DESC LIMIT 1")
	row.Scan(&response.UserID, &response.VoteID, &response.Result)
	return response
}

func GetAllResponsesForVoteID(voteID int64, db *sql.DB) (responses []Response, err error) {
	rows, err := db.Query("SELECT user_id, vote_id, result FROM responses WHERE vote_id = ?", voteID)
	if err != nil {
		return responses, err
	}
	defer rows.Close()
	for rows.Next() {
		var response Response
		rows.Scan(&response.UserID, &response.VoteID, &response.Result)
		responses = append(responses, response)
	}
	return responses, nil
}
