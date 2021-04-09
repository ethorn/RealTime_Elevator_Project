package fsm

import (
	"time"
	"fmt"
	"elevator_project/config"
)

func timer_start() { // or use golang time timer
	timerEndTime = time.Now().Add(time.Second*3)
	timerActive = true
	fmt.Println("opening door")
}

func Timer_stop() {
	timerActive = false
}

func timer_timedOut() bool {
	t := time.Now()
	duration := float64(config.DOOR_OPEN_TIME)
	return t.Sub(timerEndTime).Seconds() > duration
}

func PollTimer(doorTimeOutAlert chan<- bool) {
	for {
		switch timerActive {
		case true:
			if timer_timedOut() {
				doorTimeOutAlert <- true
			}
		default:
			time.Sleep(20*time.Millisecond)
		}
	}
}