package main

import (
	"elevator_project/cabBackup"
	"elevator_project/communication"
	"elevator_project/config"
	"elevator_project/elevio"
	"elevator_project/fsm"
	"elevator_project/network/bcast"
	"elevator_project/network/localip"
	"elevator_project/network/peers"
	"elevator_project/order_logic"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	//////////////////////////////////////// Getting flags
	var port string

	flag.StringVar(&port, "port", "15600", "port for elevator-server")

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")

	var pp int
	flag.IntVar(&pp, "pp", 2, "integer for process pair")
	flag.Parse()

	////////////////////////////////////////////////////////////////////////
	//////////////////////////////////// Process pair setup
	processPairTx := communication.ProcessPairTx
	processPairRx := communication.ProcessPairRx
	go bcast.Transmitter(58997+pp, processPairTx)
	go bcast.Receiver(58997+pp, processPairRx)

	var processPairCount int = 0
	fmt.Println("\nProcesspair: Starting as backup")

	// Try to read from primary
	for {
		if processPairCount > 500 {
			break
		}
		select {
		case <-processPairRx:
			processPairCount = 0
		default:
			time.Sleep(20 * time.Millisecond)
			processPairCount++
		}
	}
	// Become primary, start a backup, and initialize the elevator
	fmt.Println("\nProcesspair: Becoming primary and starting a backup")

	arg_pp := strconv.Itoa(pp)
	///// MAC
	currentDir, _ := os.Getwd()
	filename := "main.go"
	cmd := exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd `+currentDir+`; go run `+filename+` --id=`+id+` --port=`+port+` --pp=`+arg_pp+`"`)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	////// WINDOWS
	// arg_pp := strconv.Itoa(pp)

	// funker ikke for meg - marcus
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", `"main.go --id=`+id+` --port=`+port+` --pp=`+arg_pp+`"`).Run()
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go", "--id="+id, "--port"+port, "--pp"+arg_pp).Run()
	// err := exec.Command("cmd", "/C", "start", "powershell", "go", "run", `"pheonix.go"`).Run()
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
	stateMsgRx := communication.StateMsgRx // TODO: this should be a struct with all elevators states
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

	//////////////////////////////////// Start the elevator
	fmt.Println("Starting as slave...")
	masterCounter := 0
	for {
		// break if no master message in 6 seconds
		if masterCounter > 300 { 
			if id == currentPeers[0] {
				fmt.Println("\n... Becoming master\n")
				fsm.ElevState.Master = true

				//Redistribute orders if master disconnected
				if len(lostPeers) > 0 {
					for _, peers := range lostPeers {
						fsm.CurrentElevStates = order_logic.RedistributeOrders(fsm.CurrentElevStates, peers, fsm.ElevState.Id)
						lostPeers = fsm.RemovePeer(lostPeers, peers)
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

		case <-drv_stop:

		case e := <-doorTimeOutAlert:
			fsm.HandleDoorTimeOut(e)
			fsm.Timer_stop()

		case p := <-peerUpdateCh:
			fsm.HandlePeerUpdate(p)
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			currentPeers = p.Peers
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			lostPeers = append(lostPeers, p.Lost...)
			tempState := fsm.CurrentElevStates[p.New]
			tempState.Stuck = false
			fsm.CurrentElevStates[p.New] = tempState

			//Handle disconnected elevator as master
			if len(p.Lost) > 0 {
				if fsm.ElevState.Master {
					for _, peers := range p.Lost {
						tempState := fsm.CurrentElevStates[peers]
						tempState.Stuck = true
						fsm.CurrentElevStates[peers] = tempState
						time.Sleep(100 * time.Millisecond)
						fsm.CurrentElevStates = order_logic.RedistributeOrders(fsm.CurrentElevStates, peers, fsm.ElevState.Id)
						lostPeers = fsm.RemovePeer(lostPeers, peers)
						fsm.UpdateAllLights()
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
			if fsm.ElevState.Master {
				// First check integrity with checksum
				if communication.BtnMsgChecksumIsOk(h) {
					// Then ccknowledge message
					ackMsg := communication.AckMessage{
						MsgId:     h.Id,
						MsgSender: h.MsgSender,
					}
					ackTx <- ackMsg
				} else {
					fmt.Println("Checksum failed from hall request message: ", h.Id)
					continue
				}

				// Update CurrentElevStates with masters current state 
				fsm.CurrentElevStates[fsm.ElevState.Id] = fsm.ElevState

				// Update current Unservicable Peers
				for _, state := range fsm.CurrentElevStates {
					if state.Stuck == true {
						fsm.UnservicablePeers = append(fsm.UnservicablePeers, state.Id)
					} else {
						fsm.UnservicablePeers = fsm.RemovePeer(fsm.UnservicablePeers, state.Id)
					}
				}
				fmt.Println("Unservicable peers: ", fsm.UnservicablePeers)
				// Designate order
				index := order_logic.DesignateOrder(fsm.CurrentElevStates, h.Button, fsm.UnservicablePeers)
				designatedElev := fsm.CurrentElevStates[index]
				designatedElev.Requests[h.Button.Floor][h.Button.Button] = true
				fsm.CurrentElevStates[index] = designatedElev
				// Update states
				for _, state := range fsm.CurrentElevStates {
					statesUpdateTx <- state
				}
			}

		case s := <-statesUpdateRx:
			fsm.HandleNewElevState(s)

		case u := <-stateMsgRx:
			if communication.StateMsgChecksumIsOk(u) {
				// send acknowledge message
				ackMsg := communication.AckMessage{
					MsgId:     u.Id,
					MsgSender: u.MsgSender,
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
				}
			}
			fsm.UpdateAllLights()

		case o := <-clearOrderRx:
			if communication.ClearedOrdrMsgChecksumIsOk(o) {
				// Then Accknowledge message
				ackMsg := communication.AckMessage{
					MsgId:     o.Id,
					MsgSender: o.MsgSender,
				}
				ackTx <- ackMsg
			} else {
				fmt.Println("Checksum failed from ClearedOrder message: ", o.Id)
				continue
			}
			for _, state := range fsm.CurrentElevStates {
				state.Requests[o.Floor][elevio.BT_HallUp] = false
				state.Requests[o.Floor][elevio.BT_HallDown] = false
			}
			fsm.UpdateAllLights()

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

			// Sending alive message to the process pair backup
			processPairTx <- "Alive"
		}
	}
}
