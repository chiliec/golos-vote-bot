package main

import "testing"

func TestWaitingKey(t *testing.T) {
	userID, loginExample := 1, "chiliec"

	isWaiting, login := isWaitingKey(userID)
	if isWaiting || login != "" {
		t.Errorf("Не ждали логин %s", login)
	}

	setWaitKey(userID, loginExample)
	isWaiting, login = isWaitingKey(userID)
	if isWaiting == false || login != loginExample {
		t.Errorf("Ждём логин %s, получили %s", loginExample, login)
	}
}
