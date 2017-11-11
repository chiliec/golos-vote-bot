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
		UserID:   1,
		UserName: "chiliec",
		Rating:   10,
		Active:   true,
	}
	_, err = credential.Save(database)
	if err != nil {
		t.Fatal(err)
	}
	credentialFromDb, err := GetCredentialByUserID(credential.UserID, database)
	if err != nil {
		t.FailNow()
	}
	if credential != credentialFromDb {
		t.Errorf("\n%#v\n%#v\nНе равны!", credential, credentialFromDb)
	}
}

func TestCredential_Exists(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Failed()
	}
	credential := Credential{
		UserID:   1,
		UserName: "chiliec",
		Rating:   10,
		Active:   true,
	}
	exists := credential.Exists(database)
	if exists {
		t.Fatal("Не должно существовать")
	}
	_, err = credential.Save(database)
	if err != nil {
		t.Fatal(err)
	}
	exists = credential.Exists(database)
	if !exists {
		t.Fatal("Должно существовать")
	}
}
