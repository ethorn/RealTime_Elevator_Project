package main

import (
	"elevator_project/communication"
	"elevator_project/elevio"
	"elevator_project/fsm"
	"elevator_project/network/bcast"
	"elevator_project/network/localip"
	"elevator_project/network/peers"
	"elevator_project/order_logic"
	"elevator_project/config"
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
	fmt.Println("Starting...")
	fsm.ElevState.Id = id

	////////////////////////////////////// Init driver
	numFloors := 4 //TODO kan settes inn i config-en
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

	//////////////////////////////////// Initialize elevator
	fsm.InitElevator(id)
	if fsm.ElevState.Floor == -1 {
		fsm.OnInitBetweenFloors()
	}
	fsm.InitCurrentElevators(config.N_ELEVATORS)

	/////////////////////////////////// Wipe lights
	fsm.InitializeLights(fsm.CurrentElevStates)

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

	fsm.ElevState.Master = false // Denne er overflødig
	fmt.Println("Starting as slave...")
	masterCounter := 0
	// A peers variable with ID's of everyone connected
	var currentPeers []string
	var lostPeers []string
	var unservicablePeers []string

	for {
		// Check if master in each handler function
		//TODO hvis mulig bør denne flyttes inn i peer update-casen
		if masterCounter > 300 { // break if no master message in 6 seconds
			if id == currentPeers[0] {
				fmt.Println("... Becoming master")
				fsm.ElevState.Master = true

				//Redistribute orders if master disconnected
				if len(lostPeers) > 0 {
					for _, peers := range lostPeers {
						fsm.CurrentElevStates = order_logic.RedistributeOrders(fsm.CurrentElevStates, peers)
						if len(lostPeers) == 1 {
							lostPeers = nil
						} else {
							lostPeers = lostPeers[2:]
						}
					}
				}
			}
			masterCounter = 0
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
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			currentPeers = p.Peers
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			lostPeers = p.Lost
			unservicablePeers = p.Lost

			//Handle disconnected elevator as master
			if len(lostPeers) > 0 {
				if fsm.ElevState.Master == true {
					for _, peers := range lostPeers {
						order_logic.RedistributeOrders(fsm.CurrentElevStates, peers)
						if len(lostPeers) == 1 {
							lostPeers = nil
						} else {
							lostPeers = lostPeers[2:]
						}
					}
				}
			}

			// Update states
			if len(p.New) > 0 {
				if fsm.ElevState.Master == true {
					for _, state := range fsm.CurrentElevStates {
						statesUpdateTx <- state
					}
				}
			}

		case <-masterRx:
			masterCounter = 0

		case h := <-hallRx:
			// Only the master handles elevator orders
			if fsm.ElevState.Master == true {
				// Acknowledge order
				ackMsg := communication.AckMessage{Id: h.Id}
				ackTx <- ackMsg

				// Designate order
				fmt.Println("Unservicable peers: ", unservicablePeers)
				index := order_logic.DesignateOrder(fsm.CurrentElevStates, h.Button, unservicablePeers)
				fmt.Println(index)
				designatedElev := fsm.CurrentElevStates[index]
				// fmt.Println(designatedElev)
				designatedElev.Requests[h.Button.Floor][h.Button.Button] = true
				fmt.Println(designatedElev.Requests)
				fmt.Println("----")
				fmt.Println(designatedElev)
				fmt.Println("----")
				fsm.CurrentElevStates[index] = designatedElev

				// Update states
				for _, state := range fsm.CurrentElevStates {
					statesUpdateTx <- state
					fmt.Println("Sent state update:")
					fmt.Println(state)
				}

				// Confirm order
				elevio.SetButtonLamp(h.Button.Button, h.Button.Floor, true)
			}

		case x := <-clearOrderRx:
			elevio.SetButtonLamp(elevio.BT_HallUp, x, false)
			elevio.SetButtonLamp(elevio.BT_HallDown, x, false)

		case s := <-statesUpdateRx:
			fsm.HandleNewElevState(s)

		case u := <-stateMsgRx:
			for i, elev := range fsm.CurrentElevStates {
				if elev.Id == u.Id {
					fsm.CurrentElevStates[i] = u
					fmt.Println("Received the following state: \n", u)
				}
			}

		default:
			// For the master
			if fsm.ElevState.Master == true {
				msg := id + "is the master and is alive" // trengs ikke likevel, "Is alive" funker like fint
				masterTx <- msg
				time.Sleep(20 * time.Millisecond)
			}

			// For the slave
			masterCounter++
			// fmt.Println(masterCounter)
			time.Sleep(20 * time.Millisecond)
		}

	}

}
