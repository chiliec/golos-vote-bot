package models

import (
	"database/sql"
	"time"
)

type Vote struct {
	VoteID    int64
	UserID    int
	Author    string
	Permalink string
	Percent   int
	Completed bool
	Rejected  bool
	Date      time.Time
}

func GetVote(db *sql.DB, voteID int64) (vote Vote) {
	row := db.QueryRow("SELECT id, user_id, author, permalink, percent, completed, rejected, date " +
			   "FROM votes WHERE id = ?", voteID)
	row.Scan(&vote.VoteID, 
		 &vote.UserID, 
		 &vote.Author, 
		 &vote.Permalink, 
		 &vote.Percent, 
		 &vote.Completed, 
		 &vote.Rejected, 
		 &vote.Date)
	return vote
}

func (vote Vote) Save(db *sql.DB) (int64, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO votes(" +
		"user_id," +
		"author," +
		"permalink," +
		"percent," +
		"completed," +
		"rejected," +
		"date) " +
		"values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	result, err := prepare.Exec(vote.UserID,
				    vote.Author, 
				    vote.Permalink, 
				    vote.Percent, 
				    vote.Completed, 
				    vote.Rejected, 
				    vote.Date)
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

func GetOpenedVotesCount(db *sql.DB) (count int) {
	row := db.QueryRow("SELECT COUNT(*) FROM votes WHERE completed = 0")
	row.Scan(&count)
	return count
}

func GetTodayVotesCountForUserID(userID int, db *sql.DB) (count int) {
	row := db.QueryRow("SELECT COUNT(*) FROM votes "+
		"WHERE date > datetime('now','-1 day') AND user_id = ?", userID)
	row.Scan(&count)
	return count
}

func GetLastVote(db *sql.DB) (vote Vote) {
	row := db.QueryRow("SELECT id, user_id, author, permalink, percent, completed, rejected, date FROM votes " +
			   "ORDER BY ID DESC LIMIT 1")
	row.Scan(&vote.VoteID,
		 &vote.UserID,
		 &vote.Author,
		 &vote.Permalink,
		 &vote.Percent,
		 &vote.Completed,
		 &vote.Rejected,
		 &vote.Date)
	return vote
}

func GetAllOpenedVotes(db *sql.DB) (votes []Vote, err error) {
	rows, err := db.Query("SELECT id, user_id, author, permalink, percent, completed, rejected, date " +
			      "FROM votes WHERE completed = 0")
	if err != nil {
		return votes, err
	}
	for rows.Next() {
		var vote Vote
		rows.Scan(&vote.VoteID,
			  &vote.UserID,
			  &vote.Author,
			  &vote.Permalink,
			  &vote.Percent,
			  &vote.Completed,
			  &vote.Rejected,
			  &vote.Date)
		votes = append(votes, vote)
	}
	
	return votes, err
}

func GetOldestOpenedVote(db *sql.DB) (vote Vote) {
	row := db.QueryRow("SELECT id, user_id, author, permalink, percent, completed, rejected, date FROM votes " +
			   "WHERE completed = 0 ORDER BY date LIMIT 1")
	row.Scan(&vote.VoteID,
		 &vote.UserID,
		 &vote.Author,
		 &vote.Permalink,
		 &vote.Percent,
		 &vote.Completed,
		 &vote.Rejected,
		 &vote.Date)
	return vote
}
