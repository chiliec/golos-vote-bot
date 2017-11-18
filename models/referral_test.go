package models

import (
	"github.com/GolosTools/golos-vote-bot/db"
	"testing"
)

func TestReferral_Save(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	referral := Referral{
		UserID:    1,
		Referrer:  "chiliec",
		Completed: false,
	}
	_, err = referral.Save(database)
	if err != nil {
		t.Error(err)
	}
	referralFromDb, err := GetReferralByUserID(referral.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if referral != referralFromDb {
		t.Error("рефералы не совпадают")
	}
}

func TestReferral_SetCompleted(t *testing.T) {
	database, err := db.InitDB("")
	if err != nil {
		t.Error(err)
	}
	referral := Referral{
		UserID:    1,
		Referrer:  "chiliec",
		Completed: false,
	}
	_, err = referral.Save(database)
	if err != nil {
		t.Error(err)
	}

	referralFromDb, err := GetReferralByUserID(referral.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if referralFromDb.Completed == true {
		t.Error("реферал не должен быть completed")
	}

	referral.SetCompleted(database)
	referralFromDb2, err := GetReferralByUserID(referral.UserID, database)
	if err != nil {
		t.Error(err)
	}
	if referralFromDb2.Completed != true {
		t.Error("реферал должен быть completed")
	}
}
