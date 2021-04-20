package fsm

import (
	"elevator_project/cabBackup"
	"elevator_project/communication"
	"elevator_project/config"
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/order_logic"
	"elevator_project/single_elev_requests"
	"fmt"
	"time"
)

var d elevio.MotorDirection
var ElevState elevator.Elevator
var timerEndTime time.Time
var timerActive bool
var CurrentElevStates map[string]elevator.Elevator
var UnservicablePeers []string
var ObstructionTimer *time.Timer

func InitElevator(id string) {
	ElevState = elevator.Elevator{Id: id, Master: false, Floor: elevio.GetFloor(), Dir: elevio.MD_Stop, Behaviour: elevator.EB_Idle, Requests: cabBackup.ReadOrdersFromBackupFile(id), Stuck: false}
}

func OnInitBetweenFloors(id string) {
	ElevState.Behaviour = elevator.EB_Moving
	ElevState.Dir = elevio.MD_Down
	ElevState.Requests = cabBackup.ReadOrdersFromBackupFile(id)
	elevio.SetMotorDirection(ElevState.Dir)
}

func InitCurrentElevators(N_ELEVATORS int) {
	CurrentElevStates = make(map[string]elevator.Elevator)
	elevatorNames := []string{"one", "two", "three", "four", "five"}
	for i := 0; i < N_ELEVATORS; i++ {
		CurrentElevStates[elevatorNames[i]] = elevator.Elevator{Id: elevatorNames[i], Master: false, Floor: 0, Dir: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	}
}

func InitializeLights(CurrentElevStates map[string]elevator.Elevator) {
	for _, elev := range CurrentElevStates {
		for i, req := range elev.Requests {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, req[0])
			elevio.SetButtonLamp(elevio.BT_HallDown, i, req[1])
			elevio.SetButtonLamp(elevio.BT_Cab, i, req[2])
		}
	}
}

func RedistributeToActivePeers() {
	fmt.Println("Elevator has been stuck for too long, redistributing")
	ElevState.Stuck = true
	communication.SendStateMsg(ElevState, ElevState.Id)
	time.Sleep(100 * time.Millisecond)
	CurrentElevStates = order_logic.RedistributeOrders(CurrentElevStates, ElevState.Id, ElevState.Id)
	ElevState = CurrentElevStates[ElevState.Id]
	communication.SendStateMsg(ElevState, ElevState.Id)
}

func HandleAcknowledgeMsg(ackRx chan communication.AckMessage) {
	for {
		select {
		case ackMsg := <-ackRx:
			if ackMsg.MsgSender == ElevState.Id {
				communication.AckReceivedList.PushBack(ackMsg.MsgId)
			}
		}
	}
}

func HandleNewElevState(s elevator.Elevator) {
	UpdateAllLights()
	// Update the states of the other elevators
	for i, elev := range CurrentElevStates {
		if elev.Id == s.Id && ElevState.Id != s.Id {
			CurrentElevStates[i] = s
		}
	}

	// Update our own state (requests) if the state update it is for this elevator
	if ElevState.Id == s.Id {
		ElevState = s
		CurrentElevStates[ElevState.Id] = s

		if ElevState.Behaviour == elevator.EB_Moving {
			// If the elevator is moving, let it handle the new requests at the next floor arrival
			return
		} else if single_elev_requests.Requests_current_floor(ElevState) {
			// Handle requests at current floor
			elevio.SetDoorOpenLamp(true)
			timer_start()
			fmt.Println("Opening door")
			ElevState.Behaviour = elevator.EB_DoorOpen
			for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
				ElevState.Requests[ElevState.Floor][btn] = false
			}
			CurrentElevStates[ElevState.Id] = ElevState
			communication.SendClearedOrder(ElevState.Floor, ElevState.Id)
		} else {
			if ElevState.Behaviour == elevator.EB_DoorOpen {
				// Return and let the elevator handler new requests as the door closes.
				return
			}
			// Handle requests at other floors
			ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
			elevio.SetMotorDirection(ElevState.Dir)
			if ElevState.Dir == elevio.MD_Stop {
				ElevState.Behaviour = elevator.EB_Idle
			} else {
				ElevState.Behaviour = elevator.EB_Moving
			}
		}
		UpdateAllLights()
		communication.SendStateMsg(ElevState, ElevState.Id)
	}

}

// Handle functions
func HandleDoorTimeOut(e bool) {
	switch ElevState.Behaviour {
	case elevator.EB_DoorOpen:
		ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
		elevio.SetDoorOpenLamp(false)
		fmt.Println("Door closed.")
		elevio.SetMotorDirection(ElevState.Dir)
		if ElevState.Dir == elevio.MD_Stop {
			ElevState.Behaviour = elevator.EB_Idle
		} else {
			ElevState.Behaviour = elevator.EB_Moving
		}
		communication.SendStateMsg(ElevState, ElevState.Id)
	default:
		break
	}
	UpdateAllLights()
}

func HandleButtonEvent(btn elevio.ButtonEvent, doorTimeOutAlert chan bool) {
	fmt.Println("ButtonEvent: ", btn) // debuggin: prints out the button event

	switch ElevState.Behaviour {
	case elevator.EB_Idle:
		if ElevState.Floor == btn.Floor {
			elevio.SetDoorOpenLamp(true)
			timer_start()
			fmt.Println("Opening door")
			ElevState.Behaviour = elevator.EB_DoorOpen
			communication.SendClearedOrder(btn.Floor, ElevState.Id)
			communication.SendStateMsg(ElevState, ElevState.Id)
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
				elevio.SetMotorDirection(ElevState.Dir)
				ElevState.Behaviour = elevator.EB_Moving
				communication.SendStateMsg(ElevState, ElevState.Id)
			} else { // Hall request
				communication.SendNewHallRequest(btn, ElevState.Id)
				communication.SendStateMsg(ElevState, ElevState.Id)
			}
		}
	case elevator.EB_DoorOpen:
		if ElevState.Floor == btn.Floor {
			timer_start()
			communication.SendClearedOrder(btn.Floor, ElevState.Id)
			fmt.Println("Extending door opening timer")
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				communication.SendStateMsg(ElevState, ElevState.Id)
			} else { // hall request
				communication.SendNewHallRequest(btn, ElevState.Id)
				if ElevState.Master {
					communication.SendStateMsg(ElevState, ElevState.Id)
				}
			}
		}

	case elevator.EB_Moving:
		if btn.Button == elevio.BT_Cab {
			ElevState.Requests[btn.Floor][btn.Button] = true
			elevio.SetButtonLamp(btn.Button, btn.Floor, true)
			communication.SendStateMsg(ElevState, ElevState.Id)
		} else { // Hall request
			communication.SendNewHallRequest(btn, ElevState.Id)
			if ElevState.Master {
				communication.SendStateMsg(ElevState, ElevState.Id)
			}
		}
	}
}

func HandleNewFloor(floor int, numFloors int) {
	fmt.Println("Arriving at floor: ", floor)

	// Set new state and send it to other elevators
	ElevState.Floor = floor
	elevio.SetFloorIndicator(ElevState.Floor)

	switch ElevState.Behaviour {
	case elevator.EB_Moving: 
		if single_elev_requests.ShouldStop(ElevState) {
			// Stop the elevator and open the door
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			timer_start()
			fmt.Println("Opening door")
			ElevState.Behaviour = elevator.EB_DoorOpen

			// check which button has a running request for this floor
			for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
				if ElevState.Requests[floor][btn] {
					if btn == elevio.BT_HallUp {
						fmt.Println("entering: Hall up please")
					} else if btn == elevio.BT_HallDown {
						fmt.Println("entering: Hall down please")
					} else if btn == elevio.BT_Cab {
						fmt.Println("leaving: cab")
					}
					// clear request locally and notify other elevators
					ElevState.Requests[floor][btn] = false
					CurrentElevStates[ElevState.Id] = ElevState
				}
			}
			communication.SendClearedOrder(floor, ElevState.Id)
			communication.SendStateMsg(ElevState, ElevState.Id)
		}
	}
	UpdateAllLights()
}

func HandleChangeInObstruction(obstruction bool) {
	if obstruction {
		ObstructionTimer = time.AfterFunc(10*time.Second, RedistributeToActivePeers)
		elevio.SetMotorDirection(elevio.MD_Stop)
		extend_timer_on_obstruction()
	} else {
		ObstructionTimer.Stop()
		elevio.SetMotorDirection(d)
		timer_start()
		fmt.Println("Obstruction removed, closing door soon.")
		ElevState.Stuck = false
		communication.SendStateMsg(ElevState, ElevState.Id)
	}
}

func RemovePeer(PeerList []string, peer string) []string {
	var UpdatedPeers []string
	for _, oldpeer := range PeerList {
		if oldpeer != peer {
			UpdatedPeers = append(UpdatedPeers, oldpeer)
		}
	}
	return UpdatedPeers
}
