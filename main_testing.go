package main

import (
	"order_logic"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	//Official channels for testing information
	//ReceiveOrderChan:= make(chan ElevatorState)


	//REcieve:= make(chan ButtonEvent)

	//Variables
	//var TotalOrderList ButtonEventQueue

	// --------- Sending and recieving elevator state thorugh channel ----------
	State1 := ElevatorState{[]ButtonEvent{{2,1},{1,1}},1,1} //Example state
	State2 := ElevatorState{[]ButtonEvent{},0,-1}
	fmt.Println(State2)

	go SendElevatorState(SendElevatorStateChan, State1)
	ReceiveElevatorState(SendElevatorStateChan, &State2)
	fmt.Println(State2)
	// -------------------------- Testing accept functionality------------------------
	Order := ButtonEvent{1,1}
	var RecOrder ButtonEvent
	go AcceptOrderMaster(AcceptOrderChan, AcceptedCheck, Order)

	AcceptOrderSlave(AcceptOrderChan, AcceptedCheck, RecOrder)
	fmt.Println(RecOrder)
}
