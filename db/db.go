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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	version := getMigrationVersion(tx)
	switch version {
	case 0:
		query := `
		CREATE TABLE states(
			user_id INTEGER PRIMARY KEY NOT NULL,
			action TEXT
		);
		CREATE TABLE votes(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			user_id INTEGER,
			author TEXT,
			permalink TEXT,
			percent INTEGER,
			completed BOOLEAN NOT NULL CHECK (completed IN (0,1)) DEFAULT 0,
			date DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE credentials(
			user_id INTEGER PRIMARY KEY NOT NULL,
			user_name TEXT,
			power INTEGER NOT NULL DEFAULT 100,
			rating INTEGER NOT NULL DEFAULT 10,
			active BOOLEAN NOT NULL CHECK (active IN (0,1)) DEFAULT 0
		);
		CREATE TABLE responses(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			user_id INTEGER,
			vote_id INTEGER,
			result BOOLEAN NOT NULL CHECK (result IN (0,1)),
			date DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX idx_user_vote ON responses(user_id, vote_id);
		CREATE UNIQUE INDEX idx_author_permalink ON votes(author, permalink);
		`
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
		setMigrationVersion(tx, 1)
		fallthrough
	case 1:
		query := `
		CREATE TABLE referrals(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			user_id INTEGER UNIQUE NOT NULL,
			referrer TEXT NOT NULL,
			completed BOOLEAN NOT NULL CHECK (completed IN (0,1))
		);
		`
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
		setMigrationVersion(tx, 2)
		fallthrough
	case 3:
		query := `
		CREATE TABLE cred_temp AS (SELECT user_id, user_name, power, active FROM credentials);
		DROP TABLE credentials;
		ALTER TABLE cred_temp RENAME TO credentials;
		`
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
		setMigrationVersion(tx, 3)
		//fallthrough
	}
	tx.Commit()
	return nil
}

func getMigrationVersion(tx *sql.Tx) int {
	var version int
	row := tx.QueryRow("PRAGMA user_version")
	err := row.Scan(&version)
	if err != nil {
		log.Panic(err.Error())
	}
	return version
}

func setMigrationVersion(tx *sql.Tx, version int) {
	query := "PRAGMA user_version = " + strconv.Itoa(version)
	_, err := tx.Exec(query)
	if err != nil {
		log.Panic(err.Error())
	}
}
