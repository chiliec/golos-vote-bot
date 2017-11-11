package models

import (
	"github.com/GolosTools/golos-vote-bot/db"
	"testing"
)

func TestGetStateByUserID(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Failed()
	}
	state := State{UserID: 123, Action: "some_action"}
	_, err = state.Save(database)
	if err != nil {
		t.Fatal(err)
	}
	stateFromDatabase, err := GetStateByUserID(state.UserID, database)
	if err != nil {
		t.Fatal(err)
	}
	if state != stateFromDatabase {
		t.Fatal("Стейт не совпадает")
	}
}
