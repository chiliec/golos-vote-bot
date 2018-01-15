package models

import "database/sql"

func NewCurator(userID int, chatID int, db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT INTO referrals(" +
		"user_id," +
		"chat_id," +
		"active" +
		"values(?, ?, ?)")
	defer prepare.Close()
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(userID, chatID, true)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeactivateCurator(userID int, db *sql.DB) error {
	_, err := db.Exec("UPDATE curators SET active = 0 WHERE user_id = ?", userID)
	return err
}

func ActivateCurator(userID int, db *sql.DB) error {
	_, err := db.Exec("UPDATE curators SET active = 1 WHERE user_id = ?", userID)
	return err
}

func GetLastCuratorVotes(userID int, db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT last_votes WHERE user_id = ?", userID)
	var result *int
	err = row.Scan(&result)
	return result, err
}

func IncrementCuratorVotes(userID int, db *sql.DB) {
	_, err := db.Exec("UPDATE curators SET (total_votes, last_votes) = (SELECT total_votes, last_votes WHERE user_id = ?) WHERE user_id = ?", userID, userID)
	return err
}

func IsCuratorExists(userID int, db *sql.DB) bool {
	row := db.QueryRow("SELECT user_id FROM curators WHERE user_id = ?", userID)
	var result *int
	row.Scan(&result)
	return result != nil
}

func CleanAllLastVotes(db *sql.DB) {
	_, err := db.Exec("UPDATE curators SET last_votes = 0")
	return err	
}

func IsActiveCurator(userID int, db *sql.DB) {
	row := db.QueryRow("SELECT active FROM curators WHERE user_id = ?", userID)
	var result *int
	row.Scan(&result)
	return result	
}
