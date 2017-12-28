package Elev_control

import (
	"time"
)

var timerEndTime time.Time
var timerActive bool

func timer_start(duration time.Duration) {
	timerEndTime = time.Now().Add(duration)
	timerActive = true
}

func timer_stop() {
	timerActive = false
}

func timer_timedOut() bool {
	return (timerActive && time.Now().After(timerEndTime))
}

func GetActiveTime() int64 {
	referenceTime := time.Unix(1451606400, 0) //Unixtime 2016 01/01/00
	currentTime := time.Now()                 //returns local unix time
	return int64(currentTime.Sub(referenceTime))
}
