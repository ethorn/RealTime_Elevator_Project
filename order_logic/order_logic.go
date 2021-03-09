package order_logic

import "time"
import "sync"
import "net"
import "fmt"
import "./order_logic/elevio"


const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}


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



// ------------------------------ End self-defined structs ----------------------------------- 


func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}


func getButton(button ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

// ------------------------------------- End driver copy -------------------------------------------------


//------------------------- Support functions for the elevator order_logic -------------------------------


//-------------------------------------------- Todo ------------------------------------------------------
	// 1) First fix file-structure and 
	// 3) Function that unpacks elevator state, pulls own state w/updated queue 
	// 4) Functionality for single use elevator 
	// 5) Handling cab calls 
	// 			Cab call -> send to master for assignment -> master assigns queue with cab calls in mind ->elevator receives new queue
	// 6) Unpack and pack JSON for algo
	// 7) Poll elevatorsstates when new orders are coming in




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



func main() {
	//Official channels for testing information
	//ReceiveOrderChan:= make(chan ElevatorState)

	SendElevatorStateChan:= make(chan ElevatorState)
	AcceptOrderChan:= make(chan ButtonEvent)
	AcceptedCheck:= make(chan int)
	AcknowledgeChan:= make(chan AcknowledgeMsg)

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

