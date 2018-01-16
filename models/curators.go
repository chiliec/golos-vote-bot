package models

import "database/sql"

func NewCurator(userID int, chatID int, db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT INTO curators(" +
		"user_id," +
		"chat_id," +
		"values(?, ?, ?)")
	defer prepare.Close()
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(userID, chatID)
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

func GetLastCuratorVotes(userID int, db *sql.DB) (result int, err error) {
	row := db.QueryRow("SELECT last_votes FROM curators WHERE user_id = ?", userID)
	err = row.Scan(&result)
	return result, err
}

func IncrementCuratorVotes(userID int, db *sql.DB) error {
	row := db.QueryRow("SELECT total_votes, last_votes FROM curators WHERE user_id = ?", userID)
	var totalVotes int
	var lastVotes int
	err := row.Scan(&totalVotes, &lastVotes)
	_, err = db.Exec("UPDATE curators SET total_votes = ?, last_votes = ? WHERE user_id = ?", 
			  totalVotes+1, lastVotes+1, userID)
	return err
}

func IsCuratorExists(userID int, db *sql.DB) bool {
	row := db.QueryRow("SELECT user_id FROM curators WHERE user_id = ?", userID)
	var result *int
	err := row.Scan(&result)
	if err != nil {
		return false
	} else {
		return true
	}
}

func CleanAllLastVotes(db *sql.DB) error {
	_, err := db.Exec("UPDATE curators SET last_votes = 0")
	return err	
}

func IsActiveCurator(userID int, db *sql.DB) bool {
	row := db.QueryRow("SELECT active FROM curators WHERE user_id = ?", userID)
	var result bool
	row.Scan(&result)
	if result {
		return result	
	} else {
		return false	
	}
}

func GetAllActiveCurstorsChatID(db *sql.DB) ([]int, error) {
	var chatIDs []int
	rows, err := db.Query("SELECT chat_id FROM curators WHERE active = 1")
	if err != nil {
		return chatIDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var result int
		err = rows.Scan(&result)
		if err == nil {
			chatIDs = append(chatIDs, result)
		}
	}
	return chatIDs, err
}

func GetCuratorLastVotes(userID int, db *sql.DB) int {
	row := db.QueryRow("SELECT last_votes FROM curators WHERE user_id = ?", userID)
	var result int
	err := row.Scan(&result)
	if result != 0 && err != nil  {
		return result	
	} else {
		return 0	
	}
}

func GetAllActiveCurstorsID(db *sql.DB) ([]int, error) {
	var IDs []int
	rows, err := db.Query("SELECT user_id FROM curators WHERE active = 1")
	if err != nil {
		return IDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var result int
		err := rows.Scan(&result)
		if err == nil {
			IDs = append(IDs, result)
		}
	}
	return IDs, err
}
