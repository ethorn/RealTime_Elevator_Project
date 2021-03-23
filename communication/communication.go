package communication

import (
	"elevator_project/elevator"
	"elevator_project/elevio"
	"fmt"
	"reflect"
	"time"
)

// TODO: checksum
// TODO: message id

type AckMessage struct {
	Id int64
}

type ButtonMessage struct {
	Id int64
	Button elevio.ButtonEvent
}

// General acknowledge function
// 	- the message struct must have id as field 0 (first field)
func AcknowledgeMsg(msg interface{}, channel interface{}) {
	counter := 0

	m := reflect.ValueOf(msg)
	msgId := m.Field(0).Int()

	for {
		if counter > 150 { // Timeout at 3 sec
			break
		}
		select {
		case ack := <- AckRx:
			if ack.Id == msgId {
				fmt.Println("acknowledging true")
				return
			} else {
				AckTx <- ack
			}
		default:
			counter++
			time.Sleep(20*time.Millisecond)
		}
	}
	// Send message again
	fmt.Println("no acknowledge, sending message again")
	ch := reflect.ValueOf(channel)
	Msg := reflect.ValueOf(msg)
	ch.Send(Msg)
	// and ask for new acknowledge
	AcknowledgeMsg(msg, channel)
}

func SendStateUpdate(e elevator.Elevator) {
	StateMsgTx<-e
}

func SendNewHallRequest(btn elevio.ButtonEvent) {
	msg := ButtonMessage{Id:1, Button: btn}
	HallTx<-msg
	go AcknowledgeMsg(msg, HallTx)
}

