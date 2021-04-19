package communication

import (
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/config"
	"fmt"
	"reflect"
	"time"
	"crypto/md5"
	"strconv"
	"math/rand"
	"container/list"
)

//////////////// GLOBAL LIST FOR ACKNOWLEDGE MSG's
// TRACK RECEIVED ACKNOWLEDGE MESSAGES
var AckReceivedList = list.New()

//////////////// HELPERS
func createId() int64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int63()
}

func receivedMsgAcknowledgement(msgId int64) (bool, *list.Element) {
	for e := AckReceivedList.Front(); e != nil; e = e.Next() {
		if e.Value == msgId {
			return true, e
		}
    }
	return false, nil
}

//////////////// DEFINING MESSAGE STRUCTS
type AckMessage struct {
	MsgId int64
	MsgSender string
}

type StateMessage struct {
	Id 			int64
	Checksum	[16]byte
	State		elevator.Elevator
	MsgSender	string
}

type HallReqMessage struct {
	Id     		int64
	Checksum 	[16]byte
	Button 		elevio.ButtonEvent
	MsgSender	string
}

type ClearedOrderMessage struct {
	Id			int64
	Checksum	[16]byte
	Floor		int
	MsgSender	string
}

//////////////// CHECKSUM FUNCTIONS
// Checksum for StateMessage
func createStateChecksum(e elevator.Elevator) [16]byte {
	master 		:= ""
	floor 		:= strconv.Itoa(e.Floor)
	dir 		:= strconv.Itoa(int(e.Dir))
	behavior 	:= strconv.Itoa(int(e.Behaviour))
	requests 	:= ""

	if e.Master {
		master = "1"
	} else {
		master = "0"
	}

	for n := 0; n < config.N_FLOORS; n++ {
		for i := 0; i < config.N_BUTTONS; i++ {
			f := strconv.Itoa(n)
			b := strconv.Itoa(i)
			if e.Requests[n][i] {
				requests = requests + f + b + "1"
			} else {
				requests = requests + f + b + "0"
			}
		}
	}

	s := e.Id+master+floor+dir+behavior+requests
	md5 := md5.Sum([]byte(s))
	return md5
}

func StateMsgChecksumIsOk(msg StateMessage) bool {
	md5 := createStateChecksum(msg.State)
	if md5 == msg.Checksum {
		return true
	} else {
		return false
	}
}

// Checksum for HallReqMessage
func createHallChecksum(btn elevio.ButtonEvent) [16]byte {
	floor := strconv.Itoa(btn.Floor)
	button := strconv.Itoa(int(btn.Button))
	s := floor+button
	md5 := md5.Sum([]byte(s))
	return md5
}

func BtnMsgChecksumIsOk(msg HallReqMessage) bool {
	md5 := createHallChecksum(msg.Button)
	if md5 == msg.Checksum {
		return true
	} else {
		return false
	}
}

// Checksum for ClearedOrderMessage
func createClearedorderChecksum(floor int) [16]byte {
	f := strconv.Itoa(floor)
	md5 := md5.Sum([]byte(f))
	return md5
}

func ClearedOrdrMsgChecksumIsOk(msg ClearedOrderMessage) bool {
	md5 := createClearedorderChecksum(msg.Floor)
	if md5 == msg.Checksum {
		return true
	} else {
		return false
	}
}

func confirmAcknowledge(msg interface{}, channel interface{}, maxAckToSend int, ackSentCount int) {
	timeoutCounter := 0
	ackCounter := ackSentCount

	msgType := reflect.TypeOf(msg)
	m := reflect.ValueOf(msg)
	msgId := m.Field(0).Int()

	for {
		if timeoutCounter > 50 { // Timeout at ~1 sec
			// fmt.Println("TIMEOUT")
			break
		}
		received, e := receivedMsgAcknowledgement(msgId)
		if received {
			fmt.Println("Received acknowledge for msg: ", msgType, msgId)
			AckReceivedList.Remove(e)
			return
		} else {
			timeoutCounter++
			time.Sleep(20 * time.Millisecond)
		}
	}
	if ackCounter < maxAckToSend {
		ackCounter++
		// Send message again
		fmt.Println("no acknowledge received, resending msg: ", msgType, msgId)
		ch := reflect.ValueOf(channel)
		Msg := reflect.ValueOf(msg)
		ch.Send(Msg)
		// and ask for new acknowledge
		confirmAcknowledge(msg, channel, maxAckToSend, ackCounter)
	} else {
		fmt.Println("Too many acknowledges messages sent. Aborting trying to send this message:", msg)
		return
	}
}

// TODO endre navn
//////////////// MESSAGE SENDERS
func SendStateUpdate(e elevator.Elevator, senderId string) {
	msg := StateMessage{
		Id: 		createId(),
		Checksum: 	createStateChecksum(e),
		State:		e,
		MsgSender: 	senderId,
	}
	// StateMsgTx is available "globally" within the communication package - declared in chans.go
	StateMsgTx <- msg

	// Define max number of acknowledgements to send before aborting (and losing the message)
	// And define what to start the counter at (how many acknowledges that has been sent)
	maxAckToSend := config.MAX_ACKNOWLEDGES_TO_SEND_StateUpdate
	ackSentCount := 0
	go confirmAcknowledge(msg, StateMsgTx, maxAckToSend, ackSentCount)
}

func SendNewHallRequest(btn elevio.ButtonEvent, senderId string) {
	msg := HallReqMessage{
		Id: 		createId(),
		Checksum: 	createHallChecksum(btn), 
		Button: 	btn,
		MsgSender: 	senderId,
	}
	// HallTx is available "globally" within the communication package - declared in chans.go
	HallTx <- msg

	// Define max number of acknowledgements to send before aborting (and losing the message)
	// And define what to start the counter at (how many acknowledges that has been sent)
	maxAckToSend := config.MAX_ACKNOWLEDGES_TO_SEND_HallRequest
	ackSentCount := 0
	go confirmAcknowledge(msg, HallTx, maxAckToSend, ackSentCount)
}

func SendClearedOrder(floor int, senderId string) {
	msg := ClearedOrderMessage{
		Id: 		createId(),
		Checksum: 	createClearedorderChecksum(floor),
		Floor: 		floor,
		MsgSender: 	senderId,
	}
	// ClearedOrderTx is available "globally" within the communication package - declared in chans.go
	ClearOrderTx <- msg

	// Define max number of acknowledgements to send before aborting (and losing the message)
	// And define what to start the counter at (how many acknowledges that has been sent)
	maxAckToSend := config.MAX_ACKNOWLEDGES_TO_SEND_ClearedOrder
	ackSentCount := 0
	go confirmAcknowledge(msg, ClearOrderTx, maxAckToSend, ackSentCount)
}

