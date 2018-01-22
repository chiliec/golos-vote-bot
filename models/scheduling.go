package models

import (
	"time"
)

func WnnaSleepTill(weekday time.Weekday, hour, min int) (sleepTime time.Duration) {
	now := time.Now()
        nowYear, nowMonth, nowDay := now.Date()
	nowWeekday := now.Weekday()
	var diff int
	if nowWeekday == weekday {
		if time.Date(nowYear, nowMonth, nowDay, hour, min, 0, 0, now.Location()).Sub(time.Now()) > 0 {
			return time.Date(nowYear, nowMonth, nowDay, hour, min, 0, 0, now.Location()).Sub(time.Now())
		} else {
			diff = 7
		}
	} else { 
		diff = (7 - int(nowWeekday) + int(weekday)) % 7
	}
        nextDate := time.Date(nowYear, nowMonth, nowDay + diff, hour, min, 0, 0, now.Location())
        return nextDate.Sub(time.Now())
}

func WannaSleepOneDay(hour, min int) (sleepTime time.Duration) {
        now := time.Now()
        nowYear, nowMonth, nowDay := now.Date()
        nextDate := time.Date(nowYear, nowMonth, nowDay + 1, hour, min, 0, 0, now.Location())
        return nextDate.Sub(time.Now())
}
