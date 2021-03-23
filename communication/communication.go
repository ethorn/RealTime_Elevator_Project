package communication

import (
	"elevator_project/elevio"
	"elevator_project/elevator"
	"time"
)

// TODO: checksum
// TODO: message id

type AckMessage struct {
	Id int
}

// TODO: checksum
type ButtonMessage struct {
	Button elevio.ButtonEvent
	Id int
}

func Acknowledge() {
	
}

// TODO make an ....interface{} thing to accept all types of message channels
// TODO: cancel after X seconds?
func AcknowledgeHallMsg(msg ButtonMessage) {
	counter := 0
	for {
		if counter > 150 { // Timeout at 3 sec
			break
		}
		select {
		case a := <- AckRx:
			if a.Id == msg.Id {
				return
			} else {
				AckTx <- a
			}
		default:
			counter++
			time.Sleep(20*time.Millisecond)
		}
	}
	HallTx<-msg // Resend after timeout
	go AcknowledgeHallMsg(msg) // look for new acknowledge
}

func SendStateUpdate(e elevator.Elevator) {
	StateMsgTx<-e
}

func SendNewHallRequest(btn elevio.ButtonEvent) {
	msg := ButtonMessage{Button: btn, Id:1}
	HallTx<-msg
	go AcknowledgeHallMsg(msg)
}

