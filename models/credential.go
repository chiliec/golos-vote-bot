package models

import (
	"database/sql"
	"errors"
)

type Credential struct {
	UserID     int
	UserName   string
	PostingKey string
	Rating     int
}

func (credential Credential) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO credentials(" +
		"user_id," +
		"user_name," +
		"posting_key," +
		"rating) " +
		"values(?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(credential.UserID, credential.UserName, credential.PostingKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (credential Credential) Exists(db *sql.DB) bool {
	row := db.QueryRow("SELECT user_id FROM credentials WHERE user_id = ? AND user_name = ? AND posting_key = ?", credential.UserID, credential.UserName, credential.PostingKey)
	var userID *int
	row.Scan(&userID)
	if userID != nil {
		return true
	}
	return false
}

func (credential Credential) IncrementRating(db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET rating = rating + 1 WHERE user_id = ?", credential.UserID)
	return err
}

func (credential Credential) DecrementRating(db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET rating = rating - 1 WHERE user_id = ?", credential.UserID)
	return err
}

func (credential Credential) GetRating(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT rating FROM credentials WHERE user_id = ?", credential.UserID)
	var rating *int
	row.Scan(&rating)
	if rating != nil {
		return *rating, nil
	}
	return 0, errors.New("Не получили рейтинг из базы данных")
}

func GetAllCredentials(db *sql.DB) (credentials []Credential, err error) {
	rows, err := db.Query("SELECT user_id, user_name, posting_key, rating FROM credentials")
	if err != nil {
		return credentials, err
	}
	defer rows.Close()
	for rows.Next() {
		var credential Credential
		err := rows.Scan(&credential.UserID, &credential.UserName, &credential.PostingKey, &credential.Rating)
		if err != nil {
			return credentials, err
		}
		credentials = append(credentials, credential)
	}
	return credentials, err
}
