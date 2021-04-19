package fsm

import (
	"time"
	"fmt"
	"elevator_project/config"
)

func timer_start() {
	timerEndTime = time.Now().Add(time.Second*config.DOOR_OPEN_TIME)
	timerActive = true
}

func Timer_stop() {
	timerActive = false
}

func timer_timedOut() bool {
	t := time.Now()
	duration := float64(config.DOOR_OPEN_TIME)
	return t.Sub(timerEndTime).Seconds() > duration
}

func extend_timer_on_obstruction() {
	timerEndTime = time.Now().Add(time.Second*99999)
	timerActive = true
	fmt.Println("There is an obstruction, keeping the doop open until obstruction is removed.")
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