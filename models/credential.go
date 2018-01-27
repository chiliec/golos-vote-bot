package models

import (
	"database/sql"
	"log"
)

type Credential struct {
	UserID   int
	ChatID   int64
	UserName string
	Power    int
	Active   bool
	Curates  bool
}

func (credential Credential) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO credentials(" +
		"user_id," +
		"chat_id," +
		"user_name," +
		"power," +
		"active," +
		"curates) " +
		"values(?, ?, ?, ?, ?, ?)")
	defer prepare.Close()
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(
		credential.UserID,
		credential.ChatID,
		credential.UserName,
		credential.Power,
		credential.Active,
		credential.Curates)
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetCredentialByUserID(userID int, db *sql.DB) (credential Credential, err error) {
	row := db.QueryRow("SELECT user_id, chat_id, user_name, power, active, curates FROM credentials WHERE user_id = ?", userID)
	err = row.Scan(&credential.UserID, &credential.ChatID, &credential.UserName, &credential.Power, &credential.Active, &credential.Curates)
	return credential, err
}

func GetCredentialByUserName(userName string, db *sql.DB) (credential Credential, err error) {
	row := db.QueryRow("SELECT user_id, chat_id, user_name, power, active, curates FROM credentials WHERE user_name = ?", userName)
	err = row.Scan(&credential.UserID, &credential.ChatID, &credential.UserName, &credential.Power, &credential.Active, &credential.Curates)
	return credential, err
}

func GetAllActiveCredentials(db *sql.DB) (credentials []Credential, err error) {
	rows, err := db.Query("SELECT user_id, chat_id, user_name, power, active, curates FROM credentials")
	if err != nil {
		return credentials, err
	}
	defer rows.Close()
	for rows.Next() {
		var credential Credential
		err := rows.Scan(&credential.UserID, &credential.ChatID, &credential.UserName, &credential.Power, &credential.Active, &credential.Curates)
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

func DeactivateCurator(userID int, db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET curates = 0 WHERE user_id = ?", userID)
	return err
}

func ActivateCurator(userID int, db *sql.DB) error {
	_, err := db.Exec("UPDATE credentials SET curates = 1 WHERE user_id = ?", userID)
	return err
}

func IsActiveCurator(userID int, db *sql.DB) bool {
	row := db.QueryRow("SELECT curates FROM credentials WHERE user_id = ?", userID)
	var result bool
	row.Scan(&result)
	return result
}

func GetAllActiveCurstorsChatID(db *sql.DB) ([]int64, error) {
	var chatIDs []int64
	rows, err := db.Query("SELECT chat_id FROM credentials WHERE curates = 1")
	if err != nil {
		return chatIDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var result int64
		err = rows.Scan(&result)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		chatIDs = append(chatIDs, result)
	}
	return chatIDs, err
}

func GetAllActiveCurstorsID(db *sql.DB) ([]int, error) {
	var IDs []int
	rows, err := db.Query("SELECT user_id FROM credentials WHERE curates = 1")
	if err != nil {
		return IDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var result int
		err := rows.Scan(&result)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		IDs = append(IDs, result)
	}
	return IDs, err
}
