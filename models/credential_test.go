package models

import (
	"testing"

	"github.com/GolosTools/golos-vote-bot/db"
)

func TestDbCredentials(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Failed()
	}
	credential := Credential{
		UserID:     1,
		UserName:   "chiliec",
		PostingKey: "5...",
		Rating:     10,
	}
	credential.Save(database)
	credentialFromDb, err := GetCredentialByUserID(credential.UserID, database)
	if err != nil {
		t.FailNow()
	}
	if credential != credentialFromDb {
		t.Errorf("\n%#v\n%#v\nНе равны!", credential, credentialFromDb)
	}
}
