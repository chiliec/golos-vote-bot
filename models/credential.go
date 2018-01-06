package models

import (
	"database/sql"
	"errors"
	"strconv"
)

type Credential struct {
	UserID   int
	UserName string
	Power    int
	Rating   int
	Active   bool
}

func (credential Credential) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO credentials(" +
		"user_id," +
		"user_name," +
		"power," +
		"rating," +
		"active) " +
		"values(?, ?, ?, ?, ?)")
	defer prepare.Close()
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(
		credential.UserID,
		credential.UserName,
		credential.Power,
		credential.Rating,
		credential.Active)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (credential Credential) IncrementRating(rating int, db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET rating = rating + ? WHERE user_id = ?",
		rating, credential.UserID)
	return err
}

func (credential Credential) DecrementRating(rating int, db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET rating = rating - ? WHERE user_id = ?",
		rating, credential.UserID)
	return err
}

func (credential Credential) GetRating(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT rating FROM credentials WHERE user_id = ?",
		credential.UserID)
	var rating *int
	row.Scan(&rating)
	if rating != nil {
		return *rating, nil
	}
	return 0, errors.New("не получили рейтинг из базы данных")
}

func GetCredentialByUserID(userID int, db *sql.DB) (credential Credential, err error) {
	row := db.QueryRow("SELECT user_id, user_name, power, rating, active FROM credentials WHERE user_id = ?", userID)
	err = row.Scan(&credential.UserID, &credential.UserName, &credential.Power, &credential.Rating, &credential.Active)
	return credential, err
}

func GetCredentialByUserName(userName string, db *sql.DB) (credential Credential, err error) {
	row := db.QueryRow("SELECT user_id, user_name, power, rating, active FROM credentials WHERE user_name = ?", userName)
	err = row.Scan(&credential.UserID, &credential.UserName, &credential.Power, &credential.Rating, &credential.Active)
	return credential, err
}

func GetAllCredentials(db *sql.DB) (credentials []Credential, err error) {
	rows, err := db.Query("SELECT user_id, user_name, power, rating, active FROM credentials")
	if err != nil {
		return credentials, err
	}
	defer rows.Close()
	for rows.Next() {
		var credential Credential
		err := rows.Scan(&credential.UserID, &credential.UserName, &credential.Power, &credential.Rating, &credential.Active)
		if err == nil && credential.Active {
			credentials = append(credentials, credential)
		}
	}
	return credentials, err
}

func (credential Credential) UpdatePower(power int, db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET power = ? WHERE user_id = ?",
		power, credential.UserID)
	return err
}

func IsActiveCredential(userID int, db *sql.DB) bool {
	credential, err := GetCredentialByUserID(userID, db)
	if err != nil {
		return false
	}
	return credential.Active && len(credential.UserName) > 0
}

func CREDchangeUserID(db *sql.DB, oldID int, newID int) error {
	_, err := db.Exec("UPDATE credentials SET user_id = ? WHERE user_id = ?", newID, oldID)
	return err
}

func GetTestCredentials(db *sql.DB) (result string, err error) {
	rows, err := db.Query("SELECT user_id, user_name, power, rating, active FROM credentials WHERE user_id < 0")
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var credential Credential
		err := rows.Scan(&credential.UserID, &credential.UserName, &credential.Power, &credential.Rating, &credential.Active)
		if err == nil && credential.Active {
			result = strings.Join(result,  strconv.Itoa(credential.UserID))
		}
	}
	return result, err
}
