package order_logic

import "time"
import "sync"
import "net"
import "fmt"
import "elevator_project/elevio"

//-------------------------------------------- Todo ------------------------------------------------------
	// 1) Move "Network and message support to network-module"
	// 2) First fix file-structure and 
	// 3) Function that unpacks elevator state, pulls own state w/updated queue 
	// 4) Functionality for single use elevator 
	// 5) Handling cab calls 
	// 			Cab call -> send to master for assignment -> master assigns queue with cab calls in mind ->elevator receives new queue
	// 6) Unpack and pack JSON for algo
	// 7) Poll elevatorsstates when new orders are coming in



// --------------------------------- Start self-defined structs --------------------------------

type ElevatorState struct {
	CurrentQueue []ButtonEvent		
	LocalCabQueue []ButtonEvent		
	Dir MotorDirection				// Md_down, MD_up, MD_Stop
	Floor int						// -1 is between floors
}

type ButtonEventQueue []ButtonEvent // Needed for a function, might delete

type OrderMsg struct {  			// Struct for order flow
	order ButtonEvent
	indexElev int
	messageid int
	checksum int
}

type ElevatorMsg struct {			// Struct for elevatorstate flow
	elevators []ElevatorState
	indexElev int
	messageid int
	checksum int
}

type AcknowledgeMsg struct {		// General acknowledge message
	indexElev int
	messageid int
	checksum int
}

//-------------------------------------- General support functions ---------------------------------------

//connectedPeers := []int

//Fake function for messageid, used for testing
var messageid int = 0


func getId(lastId int) int {
	lastId = lastId + 1
	return lastId
}

//fake checksum function to test functionality
func getChecksum(data int, messageId int) int {
	data = data * messageId
	return data
}


// For testing poll functionality without using server
func PollFakeButtons(Receiver chan ButtonEvent) {
	Receiver <- ButtonEvent{1,1}
	Receiver <- ButtonEvent{1,2}
	close(Receiver)
	}


// Receives orders from pollbuttons and puts them in a unordered queue
func (OrderList *ButtonEventQueue) ReceiveButtonOrder(Receive chan ButtonEvent) {
	go PollFakeButtons(Receive)
	for n := range Receive {
        data := n 
		*OrderList = append(*OrderList, data)
    }	
}

// Functionality that supports order flow with acknowledgments between elevators
func elevatorToMsg(elevators []ElevatorState, index int, messageId int)  ElevatorMsg {
	messageId = getId(messageId)
	checksum := getChecksum(2, messageId)
	msg := ElevatorMsg{elevators, index, messageid, checksum}
	return msg
	
}

func orderToMsg(order ButtonEvent, index int, messageId int)  OrderMsg{
	messageId = getId(messageId)
	checksum := getChecksum(2, messageId)
	msg := OrderMsg{order, index, messageid, checksum}
	return msg
	
}


func ConvertToJSON(states []ElevatorState, orders ButtonEventQueue){
	//transform all info passed to the algo
}

func ConvertFromJSON( states []ElevatorState, orders ButtonEventQueue){
	//transform all info passed from the algo
}


// ------------------------------------------------- Network and message support functions ---------------------------------------------------------------

func Acknowledge(AcknowledgeChan chan AcknowledgeMsg, messageid int, connectedPeers []int, checksum int) int{ // returns 1 if info is acknowledged, returns 0 if disconnected
	tickResend := time.Tick(100 * time.Millisecond)
	tickDisconnected := time.Tick(100 * time.Millisecond)
	acceptctr := 0 
	for{
		select {
		case msg := <-AcknowledgeChan:
			if msg.messageid == messageid && msg.checksum == checksum && acceptctr != len(connectedPeers){
				acceptctr = acceptctr +1
				if acceptctr == len(connectedPeers){
					return 1
				}
			}
		case <- tickResend:
			return 0
		case <- tickDisconnected:
			return 2
		default:
			fmt.Println("Waiting")
	}
}
}

func SendOrder(Send chan OrderMsg, order OrderMsg, AcknowledgeChan chan AcknowledgeMsg) {
	Send <- order
	fmt.Println("order sent")
	acceptedStatus := Acknowledge(AcknowledgeChan, OrderMsg.messageid, connectedPeers)
	if acceptedStatus == 0{ //Loops as long as disconnected has not kicked in
		SendOrder(Send, ordermsg, AcknowledgeChan)
	}
	if acceptedStatus == 2{ //disconnected
		UpdatePeers 
		SendOrder(Send, orderMsg, AcknowledgeChan)
		
	}
	close(Send)
}

func SendOrderMaster(Send chan OrderMsg, AcknowledgeChan chan AcknowledgeMsg, Order ButtonEvent){
	ordermsg := orderToMsg(Order)
	SendOrder(Send, ordermsg, AcknowledgeChan)
	fmt.Println("Order is accepted, light on turned on")
	SetButtonLamp(Order.Button, Order.Floor, true)
}

// Pretty messy, but might be needed. 
func ReceiveOrderSlave(ReceiveOrderChan chan orderMsg, AcknowledgeChan chan AcknowledgeMsg, Order *[]ButtonEvent, elevindex int){
	order = <-ReceiveOrderChan
	answer := AcknowledgeMsg{elevindex, order.messageid, order.checksum}
	AcknowledgeChan <- answer
	fmt.Println("Slave has accepted order")
}


func SendElevatorMsg(Send chan ElevatorMsg, AcknowledgeChan chan AcknowledgeMsg, elevatormsg ElevatorMsg) {
	Send <- elevatormsg
	fmt.Println("order sent")
	acceptedStatus := Acknowledge(AcknowledgeChan, elevatormsg.messageid, connectedPeers)
	if acceptedStatus == 0{ //Loops as long as disconnected has not kicked in
		SendElevatorMsg(Send, AcknowledgeChan , elevatormsg)
	}
	if acceptedStatus == 2{ //disconnected
		UpdatePeers 
		SendElevatorMsg(Send, AcknowledgeChan , elevatormsg)
	}
	close(Send)
}

func SendStatesMaster(Send chan ElevatorMsg, AcknowledgeChan chan AcknowledgeMsg, elevators []ElevatorState){
	elevatormsg := elevatorToMsg(elevators)
	SendElevatorMsg(Send, AcknowledgeChan , elevatormsg)
	fmt.Println("Updated states are accepted")
}

func ReceiveStatesSlave(Receive chan ElevatorMsg, AcknowledgeChan chan AcknowledgeMsg, localelevators *[]ElevatorState, elevindex int){
	elevatorsmsg = <-AcceptOrderChan
	answer := AcknowledgeMsg{elevindex, elevatorsmsg.messageid, elevatorsmsg.checksum}
	AcknowledgeChan <- answer
	fmt.Println("Slave has accepted new states")
}

func GetLocalState(msg elevatormsg, index int) elevatorstate{ //Takes a message and returns the elevator state assigned to itself. 
	localState = msg.elevators[index]
	return localState
}



 // ------------------------------------------------ Single elevator support --------------------------------------------------------------

 //Still have a unsorted queue of cab and hall calls, will not take new hall-orders
func SortLocalQueue(localState ElevatorState) []ButtonEvent{
	weight_dir = 1
	weight_floordiff = 1
	orderWeightsPos := []int
	orderWeightsNeg := []int
	for buttonevent := range localState.CurrentQueue { // Have to reassign values to work with cost function
		if order.Button == 0{ orderDir:=1}
		if order.Button == 1 {orderDir:=-1}
		if order.Button == 2 {
			orderDir:=(localState.floor-order.Floor)
			if orderDir > 0 {orderDir = 1}
			if orderDir < 0 {orderDir = -1}
			if orderDir == 0 {orderDir = 0}
		}
		sortInt := weight_dir*localState.Dir*orderDir + orderDir*(localState.floor-order.Floor)^2  //Weight function for orders, subject to change. 
		if sortInt < 0{
			orderWeights.append(orderWeightsNeg, SortInt)
		}
		if sortInt == 0 || sortInt > 0 {
			orderWeights.append(orderWeightsPos, SortInt)
		}
	}
	orderWeightsPos.append(orderWeightsPos, orderWeightsNeg)
	return orderWeightsPos
}

func SingleElevatorMode(){
	ReceiveButtonOrder(ReceiveOrderChan)	
	SortLocalQueue()
	RunElevator(queue[1])
}


