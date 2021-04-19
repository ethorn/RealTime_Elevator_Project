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

	//////////////////////////////////// If we are initializing between floors
//? Hvordan funker det med importering av pakker - kjører denne etter fsm.init()?
//? Jeg får feil hvis jeg bruker denne nå?
	initialFloor := elevio.GetFloor()
	if initialFloor == -1 {
		fsm.OnInitBetweenFloors()
	}  else {
		fsm.InitElevator()
	} //TODO vurdere om evt. skal benyttes

	fsm.InitCurrentElevators(config.N_ELEVATORS)
    // Anders' implementasjon
	// {
    //     floor := GetFloor()
    //     if floor == -1 {
    //         e.Dirn = MD_Down
    //         e.Behavior = EB_Moving
    //         move <- e.Dirn
    //     } else {
    //         e.Floor = floor
    //         e.Behavior = EB_Idle
    //     }
    // }

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
	doorTimeOutAlert := make(chan bool) // TODO //? Er ikke denne done?
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go fsm.PollTimer(doorTimeOutAlert)

	/*
		How to implement "single elevator mode?"
		That is the natural way of operating
		The difference becomes when a master sends new orders
		and when new orders are sent to others
		How should elevators work while they have their own requests list?
		* internal chooseDirection algorithm
		* new cab call is added to internal requests state (and handled by internal chooseDir algo)
		* should serve cab calls when elevator is disconnected
		* This new cab call gets sent to others through updated state
		* new hall call is sent to master (with acks)
		if the master got the hall request, he checks if he is connected
		* Master then takes the hall request -> blackbox -> generates new requests for everyone. -> Send them
		Master always have the latest states of everyone, because they always send their state
		* should not serve hall requests when elev is disconnected
		* as the internal elevator gets a new state, it sends the new state to the master,
		which sends back an updated requests list (which includes cab requests)
		How to complete requests?
	*/

	fsm.ElevState.Master = false
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
                if fsm.ElevState.Master == true {
                    for _, peers := range p.Lost {
                        order_logic.RedistributeOrders(fsm.CurrentElevStates, peers)

                    }
                }
            }

            // Update states
            fmt.Print(lostPeers)
            if len(p.New) > 0 && len(lostPeers) > 0 {
                if fsm.ElevState.Master == true {
                    for _, peers := range lostPeers {
                        if len(lostPeers) == 1 {
                            lostPeers = nil
                        } else {
                            lostPeers = lostPeers[2:]
                        }
                        fmt.Print("Updated cab calls of elevator", peers, "\n")
                    }
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
				designatedelev := fsm.CurrentElevStates[index]
				designatedelev.Requests[h.Button.Floor][h.Button.Button] = true
				fsm.CurrentElevStates[index] = designatedelev

				// Update states
				for _, state := range fsm.CurrentElevStates {
					statesUpdateTx <- state
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
					fmt.Println("Received the following state: \n", u) //TODO denne blir fort veldig overveldet
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
