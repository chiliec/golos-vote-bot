package db

import "testing"

func TestInitDB(t *testing.T) {
	db, err := InitDB("")
	if err != nil {
		t.Error(err)
	}
	err = db.Close()
	if err != nil {
		t.Error(err)
	}
}
