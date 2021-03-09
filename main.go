package main

import "elevator_project/network_driver/bcast"
import (
	//"/network_driver/bcast"
	//"./network_driver/conn"
	//"./network_driver/localip"
	//"./network_driver/peers"
	//"./network_core"
	//"./order_logic"
	// "flag"
	"fmt"
	// "os"
	// "time"
)

func main() {
	transmitChan := make(chan int)
	go bcast.Transmitter(123456, transmitChan)
	fmt.Println("hej")
}

// func main() {
// 	//Channels
// 		ReceiveOrderChan:= make(chan ElevatorState)
// 		AcceptOrderChan:= make(chan ButtonEvent)
// 		AcceptedCheck:= make(chan int)
		
// 		// SendElevatorStateChan:= make(chan ElevatorState)
// 		// AcceptOrderChan:= make(chan ButtonEvent)
// 		// AcknowledgeChan:= make(chan AcknowledgeMsg)

// 	// todo var PriorityQueue

// 	if priority == 1{

// 	for{

// 		select{

// 		case Order <- ReceiveOrderChan:
// 			AcceptOrderMaster(AcceptOrderChan chan ButtonEvent,  AcceptedCheck chan int, Order ButtonEvent)(true)
// 			if AcceptOrderMaster == true{
// 				//
// 				//Pass current queue, []Elevatorstate(Elev1, Elev2, Elev3) states to algo
// 			}

// 		case i <- SendElevatorStateChan:
// 			// []Elevatorstate(Elev1, Elev2, Elev3)

// 		case i <- OrderFulfilled:
// 			// Delete order from active queue

// 		case CheckConnection() == 0:
// 			//do something
		



// 		default: 
// 			ReceiveButtonOrder(ReceiveOrderChan)	
// 			SendElevatorState(Send chan ElevatorState, state ElevatorState)
// 			CheckConnection()
// 			PullOwnQueue()
// 			RunElevator(queue[1])

// 		}
// 	}
// }



// //Slave implementation
// else if priority == 0{
// 	for{

// 		select{

// 		case i <- SendElevatorStateChan:
// 			// []Elevatorstate(Elev1, Elev2, Elev3)

// 		case i <- ReceiveElevatorState(Receive chan ElevatorState, state *ElevatorState):
// 			// Delete order from active queue

// 		case CheckConnection() == 0:
// 			//do something

// 		default: 
// 			ReceiveCallButtonOrder()	
// 			SendElevatorState(Send chan ElevatorState, state ElevatorState)
// 			CheckConnection()
// 			RunElevator(queue[1])
// 		}
// 	}	
// }


// else if disconnected == 1{
// 	for{
// 		select{
// 			case <- connected:
// 				ErrorHandlerReconnect()
// 					//Pass cab calls remaining
// 					//Receive updated state from master and overwrite
// 				disconnected = 0

// 			default: 
// 			ReceiveCallButtonOrder() //Cab calls are the only calls being served + queue at disconnect
// 			BasicOrderDesignator()
// 			RunElevator(queue[1])

// }