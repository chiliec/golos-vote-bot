package models

import (
	"testing"

	"github.com/GolosTools/golos-vote-bot/db"
)

func TestDbCredentials(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	credential := Credential{
		UserID:   1,
		UserName: "chiliec",
		Power:    100,
		Active:   true,
	}
	_, err = credential.Save(database)
	if err != nil {
		t.Error(err)
	}
	credentialFromDb, err := GetCredentialByUserID(credential.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if credential != credentialFromDb {
		t.Errorf("\n%#v\n%#v\nНе равны!", credential, credentialFromDb)
	}
}

func TestCredential_UpdatePower(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	credential := Credential{
		UserID:   1,
		UserName: "chiliec",
		Power:    100,
		Active:   true,
	}
	_, err = credential.Save(database)
	if err != nil {
		t.Error(err)
	}
	credentialFromDB, err := GetCredentialByUserID(credential.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if credentialFromDB.Power != credential.Power {
		t.Error("Начальная сила неправильная")
	}

	err = credential.UpdatePower(42, database)
	if err != nil {
		t.Error(err)
	}
	updatedCredential, err := GetCredentialByUserID(credential.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if updatedCredential.Power != 42 {
		t.Error("Сила не обновилась")
	}
}

func TestIsActiveCredential(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	credential := Credential{
		UserID:   1,
		UserName: "chiliec",
		Power:    100,
		Active:   true,
	}
	exists := IsActiveCredential(credential.UserID, database)
	if exists {
		t.Error("Не должен быть активным")
	}
	_, err = credential.Save(database)
	if err != nil {
		t.Error(err)
	}
	exists = IsActiveCredential(credential.UserID, database)
	if !exists {
		t.Error("Должен существовать")
	}
}
