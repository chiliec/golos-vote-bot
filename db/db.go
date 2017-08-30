package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strconv"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return db, err
	}
	err = createTables(db)
	return db, err
}

func createTables(db *sql.DB) error {
	version := getMigrationVersion(db)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	switch version {
	case 0:
		query := `
		CREATE TABLE votes(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			user_id INTEGER,
			author TEXT,
			permalink TEXT,
			percent INTEGER
		);
		`
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
		setMigrationVersion(db, 1)
		fallthrough
	case 1:
		query := `
		CREATE TABLE credentials(
			user_id INTEGER PRIMARY KEY NOT NULL,
			user_name TEXT,
			posting_key TEXT
		);
		`
		_, err := db.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
		setMigrationVersion(db, 2)
		//fallthrough
	}
	tx.Commit()
	return nil
}

func getMigrationVersion(db *sql.DB) int {
	var version int
	row := db.QueryRow("PRAGMA user_version")
	err := row.Scan(&version)
	if err != nil {
		log.Panic(err.Error())
	}
	return version
}

func setMigrationVersion(db *sql.DB, version int) {
	query := "PRAGMA user_version = " + strconv.Itoa(version)
	_, err := db.Exec(query)
	if err != nil {
		log.Panic(err.Error())
	}
}
