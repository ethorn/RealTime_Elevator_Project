package main

import (
	"elevator_project/communication"
	"elevator_project/elevio"
	"elevator_project/fsm"
	"elevator_project/network/bcast"
	"elevator_project/network/localip"
	"elevator_project/network/peers"
	"elevator_project/order_logic"
	"elevator_project/cabBackup"
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
	// numFloors := 4 //! Brukte heller config-en
	fmt.Println("localhost:" + port)
	elevio.Init("localhost:"+port, config.N_FLOORS)

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
	peerTxEnable := communication.PeerTxEnable
	go peers.Transmitter(45647, id, peerTxEnable)
	go peers.Receiver(45647, peerUpdateCh)

	// make channels for ...
	stateMsgTx := communication.StateMsgTx
	stateMsgRx := communication.StateMsgRx  // TODO: this should be a struct with all elevators states
	go bcast.Transmitter(46569, stateMsgTx)
	go bcast.Receiver(46569, stateMsgRx)

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
	go fsm.HandleAcknowledgeMsg(ackRx)

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
	doorTimeOutAlert := make(chan bool)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go fsm.PollTimer(doorTimeOutAlert)

	//////////////////////////////////// Initialize cab backup module
	cabBackup.Init(id)	

	//////////////////////////////////// Initialize elevator
	
	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors(id)
	} else {
		fsm.InitElevator(id)
	}
	fsm.InitCurrentElevators(config.N_ELEVATORS)
	fsm.InitializeLights(fsm.CurrentElevStates)

	//////////////////////////////////// Keeping track of peers
	var currentPeers []string
	var lostPeers []string
	var unservicablePeers []string

	//////////////////////////////////// Start the elevator
	fmt.Println("Starting as slave...")
	masterCounter := 0
	for {
		// Check if master in each handler function
		if masterCounter > 300 { // break if no master message in 6 seconds
			if id == currentPeers[0] {
				fmt.Println("\n... Becoming master\n")
				fsm.ElevState.Master = true

				//Redistribute orders if master disconnected
				if len(lostPeers) > 0 {
					for _, peers := range lostPeers {
						fsm.CurrentElevStates = order_logic.RedistributeOrders(fsm.CurrentElevStates, peers, fsm.ElevState.Id)
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
			fsm.HandleNewFloor(b, config.N_FLOORS)

		case c := <-drv_obstr:
			fsm.HandleChangeInObstruction(c)

		case d := <-drv_stop:
			fsm.HandleChangeInStopBtn(d, config.N_FLOORS)

		case e := <-doorTimeOutAlert:
			fsm.HandleDoorTimeOut(e)
			fsm.Timer_stop()
		
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
            fmt.Printf("  Peers:    %q\n", p.Peers)
            currentPeers = p.Peers
            fmt.Printf("  New:      %q\n", p.New)
            fmt.Printf("  Lost:     %q\n", p.Lost)
            lostPeers = append(lostPeers, p.Lost...)
            unservicablePeers = p.Lost

            //Handle disconnected elevator as master
			if len(p.Lost) > 0 {
                if fsm.ElevState.Master {
                    for _, peers := range p.Lost {
                        fsm.CurrentElevStates = order_logic.RedistributeOrders(fsm.CurrentElevStates, peers, fsm.ElevState.Id)
                        lostPeers = fsm.RemovePeer(lostPeers, peers)
                    }
                }
            }

			if len(p.New) > 0 && id == p.New {
				state := fsm.ElevState // Duplicate in order to avoid mutation
				statesUpdateTx <- state
			}

		case <-masterRx:
			masterCounter = 0

		case h := <-hallRx:
			// Only the master handles elevator orders
			if fsm.ElevState.Master {
				// First check integrity with checksum
				if communication.BtnMsgChecksumIsOk(h) {
					// Then ccknowledge message
					ackMsg := communication.AckMessage{
						MsgId: 		h.Id,
						MsgSender: 	h.MsgSender,
					}
					ackTx <- ackMsg
				} else {
					fmt.Println("Checksum failed from hall request message: ", h.Id)
					continue
				}

				// Update with masters current state
				fsm.CurrentElevStates[fsm.ElevState.Id] = fsm.ElevState

				// Designate order
				fmt.Println("Unservicable peers: ", unservicablePeers)
				index := order_logic.DesignateOrder(fsm.CurrentElevStates, h.Button, unservicablePeers)
				designatedElev := fsm.CurrentElevStates[index]
				designatedElev.Requests[h.Button.Floor][h.Button.Button] = true
				fsm.CurrentElevStates[index] = designatedElev

				// Update states
				for _, state := range fsm.CurrentElevStates {
					statesUpdateTx <- state
				}

				// Confirm order
				elevio.SetButtonLamp(h.Button.Button, h.Button.Floor, true)
			}

		case o := <-clearOrderRx:
			// First check integrity with checksum
			if communication.ClearedOrdrMsgChecksumIsOk(o) {
				// Then Accknowledge message
				ackMsg := communication.AckMessage{
					MsgId: 		o.Id,
					MsgSender: 	o.MsgSender,
				}
				ackTx <- ackMsg
			} else {
				fmt.Println("Checksum failed from ClearedOrder message: ", o.Id)
				continue
			}
//			fmt.Println("-- Received confirmation of order and turning off lamp: ", o.Id)
			for _, state := range fsm.CurrentElevStates {
				state.Requests[o.Floor][elevio.BT_HallUp] = false
				state.Requests[o.Floor][elevio.BT_HallDown] = false
			}
			elevio.SetButtonLamp(elevio.BT_HallUp, o.Floor, false)
			elevio.SetButtonLamp(elevio.BT_HallDown, o.Floor, false)

		case s := <-statesUpdateRx:
			fsm.HandleNewElevState(s)

		case u := <-stateMsgRx:
			// First check integrity with checksum
			if communication.StateMsgChecksumIsOk(u) {
				// Then ccknowledge message
				ackMsg := communication.AckMessage{
					MsgId: 		u.Id,
					MsgSender: 	u.MsgSender,
				}
				ackTx <- ackMsg
			} else {
				fmt.Println("Checksum failed from state message: ", u.Id)
				continue
			}

			if fsm.ElevState.Id == u.State.Id {
				cabBackup.WriteOrdersToBackupFile(u.State.Requests)
				continue
			}

			for i, elev := range fsm.CurrentElevStates {
				if elev.Id == u.State.Id {
					fsm.CurrentElevStates[i] = u.State
					// currentRequests := fsm.CurrentElevStates[i].Requests // Duplicate in order to avoid mutation, probably redundant
					// cabBackup.WriteOrdersToBackupFile(currentRequests)
				}
			}

		default:
			// For the master
			if fsm.ElevState.Master {
				msg := "alive"
				masterTx <- msg
				time.Sleep(20 * time.Millisecond)
			}
			// For the slave
			masterCounter++
			time.Sleep(20 * time.Millisecond)
		}

	}

}
