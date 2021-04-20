package communication

import (
	"elevator_project/elevator"
	"elevator_project/network/peers"
)
// Make channels for handling the process pair communication
var ProcessPairTx chan string = make(chan string)
var ProcessPairRx chan string = make(chan string)

// Make channels for handling peer connection updates
var PeerUpdateCh chan peers.PeerUpdate = make(chan peers.PeerUpdate)
var PeerTxEnable chan bool = make(chan bool)

// make channels for sending and receiving our custom data types
// StateMessage is declared in communication.go
var StateMsgTx chan StateMessage = make(chan StateMessage)
var StateMsgRx chan StateMessage = make(chan StateMessage)

// Make channels for sending and receiving states which should update the slave states
var StatesUpdateTx chan elevator.Elevator = make(chan elevator.Elevator)
var StatesUpdateRx chan elevator.Elevator = make(chan elevator.Elevator)

// Make channels for sending/receiving hall requests
var HallTx chan HallReqMessage = make(chan HallReqMessage)
var HallRx chan HallReqMessage = make(chan HallReqMessage)

// Make channels for sending and receiving acknowledgements between elevators
var AckTx chan AckMessage = make(chan AckMessage)
var AckRx chan AckMessage = make(chan AckMessage)

// Internal acknowledge channel
var AcknowledgeChan chan AckMessage = make(chan AckMessage)

// Make channels for sending/recieving Master/Slave messages
var MasterTx chan string = make(chan string, 1)
var MasterRx chan string = make(chan string, 1)

// Make channels for sending/recieving Order Updates
var ClearOrderTx chan ClearedOrderMessage = make(chan ClearedOrderMessage)
var ClearOrderRx chan ClearedOrderMessage = make(chan ClearedOrderMessage)
