package models

import (
	"database/sql"
	"time"
)

func GetLastReportDate(db *sql.DB) (lastReportDate time.Time) {
	row := db.QueryRow("SELECT date FROM events WHERE type = 'POST' ORDER BY date DESC LIMIT 1")
	row.Scan(&lastReportDate)
	return lastReportDate
}

func NewReportPosted(db *sql.DB)  
