package models

import (
	"database/sql"
	"time"
)

function WannaSleepOneDay(hour, min int) (sleepTime time.Duration) {
        now := time.Now()
        nowYear, nowMonth, nowDay := now.Date()
        nextDate := time.Date(nowYear, nowMonth, nowDay + 1, hour, min, 0, 0, now.Location())
        return nextDate.Sub(time.Now())
}
