package models

import (
	"database/sql"
	"time"
)

func GetLastRewardDate(db *sql.DB) (lastReportDate time.Time) {
	row := db.QueryRow("SELECT date FROM events WHERE type = 'REWARD' ORDER BY date DESC LIMIT 1")
	row.Scan(&lastReportDate)
	return lastReportDate
}

func NewRewardDistributed(db *sql.DB) (int) {
	result, err := db.Exec("INSERT INTO events (type) VALUES ('REWARD')")
	return result.LastInsertId()
}
