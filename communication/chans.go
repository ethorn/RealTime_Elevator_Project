package communication

import (
	"elevator_project/elevator"
	"elevator_project/network/peers"
)

var PeerUpdateCh chan peers.PeerUpdate = make(chan peers.PeerUpdate)
var PeerTxEnable chan bool = make(chan bool)

// make channels for sending and receiving our custom data types
var StateMsgTx chan elevator.Elevator = make(chan elevator.Elevator)
var StateMsgRx chan elevator.Elevator = make(chan elevator.Elevator)

// Make channels for sending and receiving states which should update the slave states
var StatesUpdateTx chan elevator.Elevator = make(chan elevator.Elevator)
var StatesUpdateRx chan elevator.Elevator = make(chan elevator.Elevator)

// Make channels for sending/receiving hall requests
var HallTx chan ButtonMessage = make(chan ButtonMessage)
var HallRx chan ButtonMessage = make(chan ButtonMessage)

// Make channels for acknowledgements
var AckTx chan AckMessage = make(chan AckMessage)
var AckRx chan AckMessage = make(chan AckMessage)

// Make channels for sending/recieving Master/Slave messages
var MasterTx chan string = make(chan string, 1)
var MasterRx chan string = make(chan string, 1)
