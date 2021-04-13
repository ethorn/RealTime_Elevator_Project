package main

import (
	"elevator_project/communication"
	"elevator_project/elevio"
	"elevator_project/fsm"
	"elevator_project/network/bcast"
	"elevator_project/network/localip"
	"elevator_project/network/peers"
	"elevator_project/order_logic"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	//////////////////////////////////////// Getting flags
	var port string

	flag.StringVar(&port, "port", "15600", "port for elevator-server")

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	//////////////////////////////////////// State machine initialization
	fmt.Println("Starting..")
	fsm.ElevState.Id = id
	//TODO
	// Check if we are initializing between floors
	// if elevio.getFloor() == -1 {

	// }
	// if so, fsm.onInitBetweenFloors

	////////////////////////////////////// Init driver
	numFloors := 4
	fmt.Println("localhost:" + port)
	elevio.Init("localhost:"+port, numFloors)

	//////////////////////////////////// Init UDP broadcast channels
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerUpdateCh := communication.PeerUpdateCh
	// We can disable/enable the transmitter after it has been started. This could be used to signal that we are somehow "unavailable".
	peerTxEnable := communication.PeerTxEnable
	go peers.Transmitter(45647, id, peerTxEnable)
	go peers.Receiver(45647, peerUpdateCh)

	// make channels for sending and receiving our custom data types
	stateMsgTx := communication.StateMsgTx
	stateMsgRx := communication.StateMsgRx  // TODO: this should be a struct with all elevators states
	go bcast.Transmitter(46569, stateMsgTx) // These functions can take any number of channels!
	go bcast.Receiver(46569, stateMsgRx)    // It is also possible to start multiple transmitters/receivers on the same port.

	// Make channels for sending and receiving states which should update the slave states
	statesUpdateTx := communication.StatesUpdateTx
	statesUpdateRx := communication.StatesUpdateRx
	go bcast.Transmitter(61569, statesUpdateTx)
	go bcast.Receiver(61569, statesUpdateRx)

	// Make channels for sending/receiving hall requests
	hallTx := communication.HallTx
	hallRx := communication.HallRx
	go bcast.Transmitter(49731, hallTx)
	go bcast.Receiver(49731, hallRx)

	// Make channels for acknowledgements
	ackTx := communication.AckTx
	ackRx := communication.AckRx
	go bcast.Transmitter(41932, ackTx)
	go bcast.Receiver(41932, ackRx)

	// Make channels for sending/recieving Master/Slave messages
	masterTx := communication.MasterTx
	masterRx := communication.MasterRx
	go bcast.Transmitter(58989, masterTx)
	go bcast.Receiver(58989, masterRx)

	clearOrderTx := communication.ClearOrderTx
	clearOrderRx := communication.ClearOrderRx
	go bcast.Transmitter(58990, clearOrderTx)
	go bcast.Receiver(58990, clearOrderRx)

	////////////////////////////////// initialize I/O channels polling
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	doorTimeOutAlert := make(chan bool) // TODO
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go fsm.PollTimer(doorTimeOutAlert)

	// How to implement "single elevator mode?"
	// That is the natural way of operating
	// The difference becomes when a master sends new orders
	// and when new orders are sent to others
	// How should elevators work while they have their own requests list?
	// * internal chooseDirection algorithm
	// * new cab call is added to internal requests state (and handled by internal chooseDir algo)
	// * should serve cab calls when elevator is disconnected
	// * This new cab call gets sent to others through updated state
	// * new hall call is sent to master (with acks)
	// if the master got the hall request, he checks if he is connected
	// * Master then takes the hall request -> blackbox -> generates new requests for everyone. -> Send them
	// Master always have the latest states of everyone, because they always send their state
	// * should not serve hall requests when elev is disconnected
	// * as the internal elevator gets a new state, it sends the new state to the master,
	// which sends back an updated requests list (which includes cab requests)
	// How to complete requests?

	// Use same handle functions, just use an if statement to see if the elevator is master or not

	fmt.Println("Starting as slave...")
	masterCounter := 0
	for {
		if masterCounter > 300 { // break if no master message in 6 seconds
			break
		}
		select {
		case a := <-drv_buttons:
			fsm.HandleButtonEvent(a, doorTimeOutAlert)
		case b := <-drv_floors:
			fsm.HandleNewFloor(b, numFloors)
		case c := <-drv_obstr:
			fsm.HandleChangeInObstruction(c)
		case d := <-drv_stop:
			fsm.HandleChangeInStopBtn(d, numFloors)
		case e := <-doorTimeOutAlert:
			fsm.HandleDoorTimeOut(e)
			fsm.Timer_stop()
		case p := <-peerUpdateCh:
			// TODO
			// Set a peers variable with ID's of everyone connected
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case <-masterRx:
			masterCounter = 0
			// fmt.Println(m)
		case s := <-statesUpdateRx:
			fsm.HandleNewElevState(s) // TODO
		case u := <-clearOrderRx:
			elevio.SetButtonLamp(elevio.BT_HallUp, u, false)
			elevio.SetButtonLamp(elevio.BT_HallDown, u, false)
		default:
			masterCounter++
			// fmt.Println(masterCounter)
			time.Sleep(20 * time.Millisecond)
		}
	}

	fmt.Println("...Becoming master")
	fsm.ElevState.Master = true
	for {
		select {
		// Check if master in each handler function
		case a := <-drv_buttons:
			fsm.HandleButtonEvent(a, doorTimeOutAlert)
		case b := <-drv_floors:
			fsm.HandleNewFloor(b, numFloors)
		case c := <-drv_obstr:
			fsm.HandleChangeInObstruction(c)
		case d := <-drv_stop:
			fsm.HandleChangeInStopBtn(d, numFloors)
		case e := <-doorTimeOutAlert:
			fsm.HandleDoorTimeOut(e)
			fsm.Timer_stop()
		case p := <-peerUpdateCh:
			// TODO
			// Set a peers variable with ID's of everyone connected
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case h := <-hallRx:

			// Acknowledge order
			ackMsg := communication.AckMessage{Id: h.Id}
			ackTx <- ackMsg

			// Designate order
			index := order_logic.DesignateOrder(fsm.CurrentElevStates, h.Button)
			fsm.CurrentElevStates[index].Requests[h.Button.Floor][h.Button.Button] = true

			// If self-designated
			if fsm.CurrentElevStates[index].Id == fsm.ElevState.Id {
				fsm.ElevState = fsm.CurrentElevStates[index]
				fsm.CurrentElevStates[index].Requests[h.Button.Floor][h.Button.Button] = false
				fsm.HandleNewElevState(fsm.ElevState)
			}

			// For all connected peers - TODO: Mutable
			for i := 0; i < 3; i++ {
				state := fsm.CurrentElevStates[i]
				statesUpdateTx <- state
			}
			// Confirm order
			elevio.SetButtonLamp(h.Button.Button, h.Button.Floor, true)

		case u := <-clearOrderRx:
			elevio.SetButtonLamp(elevio.BT_HallUp, u, false)
			elevio.SetButtonLamp(elevio.BT_HallDown, u, false)

		case s := <-stateMsgRx:
			for i, elev := range fsm.CurrentElevStates {
				if elev.Id == s.Id {
					fsm.CurrentElevStates[i] = s
				}
			}

		default:
			// fmt.Println("sending message")
			msg := "Alive"
			masterTx <- msg
			time.Sleep(20 * time.Millisecond)
		}
	}
}
