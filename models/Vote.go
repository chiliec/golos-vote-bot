package models

import (
	"database/sql"
	"errors"
)

type Vote struct {
	Voter     string
	Author    string
	Permalink string
	Percent   int
}

func (vote Vote) Save(db *sql.DB, userID int) (bool, error) {
	prepare, err := db.Prepare("INSERT INTO votes(" +
		"user_id," +
		"author," +
		"permalink," +
		"percent) " +
		"values(?, ?, ?, ?)")
	if err != nil {
		return false, errors.New("Что-то не так при подготовке сохранения модели")
	}
	_, err = prepare.Exec(userID, vote.Author, vote.Permalink, vote.Percent)
	if err != nil {
		return false, errors.New("Не смогли сохранить модель")
	}
	return true, nil
}
