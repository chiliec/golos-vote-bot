package models

import "database/sql"

type State struct {
	UserID int
	Action string
}

func (state State) Save(db *sql.DB) (bool, error) {
	prepare, err := db.Prepare("INSERT OR REPLACE INTO states(" +
		"user_id," +
		"action) " +
		"values(?, ?)")
	if err != nil {
		return false, err
	}
	_, err = prepare.Exec(state.UserID, state.Action)
	return err != nil, err
}

func GetStateByUserID(userID int, db *sql.DB) (state State, err error) {
	row := db.QueryRow("SELECT user_id, action FROM states WHERE user_id = ?", userID)
	err = row.Scan(&state.UserID, &state.Action)
	if err != nil {
		if err == sql.ErrNoRows {
			return State{UserID: userID, Action: ""}, nil
		}
	}
	return state, err
}
