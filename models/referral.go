package models

import "database/sql"

type Referral struct {
	UserID    int
	Referrer  string
	Completed bool
}

func (referral Referral) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT INTO referrals(" +
		"user_id," +
		"referrer," +
		"completed) " +
		"values(?, ?, ?)")
	defer prepare.Close()
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(
		referral.UserID,
		referral.Referrer,
		referral.Completed)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (referral Referral) SetCompleted(db *sql.DB) error {
	_, err := db.Exec("UPDATE referrals SET completed = 1 WHERE user_id = ?", referral.UserID)
	return err
}

func GetReferralByUserID(userID int, db *sql.DB) (referral Referral, err error) {
	row := db.QueryRow("SELECT user_id, referrer, completed FROM referrals WHERE user_id = ?", userID)
	err = row.Scan(&referral.UserID, &referral.Referrer, &referral.Completed)
	return referral, err
}

func REFchangeUserID(db *sql.DB, oldID int, newID int) error {
	_, err := db.Exec("UPDATE referrals SET user_id = ? WHERE user_id = ?", newID, oldID)
	return err
}
