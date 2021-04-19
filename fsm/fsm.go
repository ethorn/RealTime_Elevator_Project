package fsm

import (
	"elevator_project/communication"
	"elevator_project/config"
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/single_elev_requests"
	"fmt"
	"time"
)

// IMPORTANT: ALL STATE CHANGES HAPPEN IN THIS FILE

var d elevio.MotorDirection
var ElevState elevator.Elevator
var timerEndTime time.Time
var timerActive bool
var CurrentElevStates map[string]elevator.Elevator

func InitElevator(id string) {
	ElevState = elevator.Elevator{Id: id, Master: false, Floor: elevio.GetFloor(), Dir: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	communication.SendStateUpdate(ElevState, ElevState.Id)
}

func OnInitBetweenFloors(id string) {
	ElevState.Behaviour = elevator.EB_Moving
	ElevState.Dir = elevio.MD_Down
	elevio.SetMotorDirection(ElevState.Dir)
}

//? Hvorfor Floor: 0?
func InitCurrentElevators(N_ELEVATORS int) {
	// rename to AllElevStates?
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

func UpdateLights(s elevator.Elevator) {
	for i, call := range s.Requests {
		if call[0] {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
		}
		if call[1] {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
		}
		if s.Id == ElevState.Id {
			if call[2] {
				elevio.SetButtonLamp(elevio.BT_Cab, i, true)
			}
		}
	}
}

func HandleAcknowledgeMsg(ackRx chan communication.AckMessage) {
	for {
		select {
		case ackMsg := <- ackRx:
			if ackMsg.MsgSender == ElevState.Id {
				communication.AckReceivedList.PushBack(ackMsg.MsgId)
			}
		}
	}
}

func HandleNewElevState(s elevator.Elevator) {
	fmt.Println("--Got new state and updating lights")
	UpdateLights(s)
	// Update the states of the other elevators
	for i, elev := range CurrentElevStates {
		if elev.Id == s.Id &&  ElevState.Id != s.Id{
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
			fmt.Println(ElevState)
			ElevState.Behaviour = elevator.EB_DoorOpen
			for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
				ElevState.Requests[ElevState.Floor][btn] = false
			}
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
		communication.SendStateUpdate(ElevState, ElevState.Id)
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
		communication.SendStateUpdate(ElevState, ElevState.Id)
	default:
		break
	}
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
			communication.SendStateUpdate(ElevState, ElevState.Id)
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
				elevio.SetMotorDirection(ElevState.Dir)
				ElevState.Behaviour = elevator.EB_Moving
				communication.SendStateUpdate(ElevState, ElevState.Id)
			} else { // Hall request
				communication.SendNewHallRequest(btn, ElevState.Id)
				communication.SendStateUpdate(ElevState, ElevState.Id)
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
				communication.SendStateUpdate(ElevState, ElevState.Id)
			} else { // hall request
				communication.SendNewHallRequest(btn, ElevState.Id)
				if ElevState.Master {
					communication.SendStateUpdate(ElevState, ElevState.Id)
				}
			}
		}

	case elevator.EB_Moving:
		if btn.Button == elevio.BT_Cab {
			ElevState.Requests[btn.Floor][btn.Button] = true
			elevio.SetButtonLamp(btn.Button, btn.Floor, true)
			communication.SendStateUpdate(ElevState, ElevState.Id)
		} else { // Hall request
			communication.SendNewHallRequest(btn, ElevState.Id)
			if ElevState.Master {
				communication.SendStateUpdate(ElevState, ElevState.Id)
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
					elevio.SetButtonLamp(btn, floor, false)
				}					
			}
			communication.SendClearedOrder(floor, ElevState.Id)
			communication.SendStateUpdate(ElevState, ElevState.Id)
		}
	}
}

func HandleChangeInObstruction(obstruction bool) {
	fmt.Println("Obstruction?", obstruction) //? Erstatter tonnevis av printf's med println's, går vel greit kodemessig?
	if obstruction {
		elevio.SetMotorDirection(elevio.MD_Stop)
		extend_timer_on_obstruction()
	} else {
		elevio.SetMotorDirection(d)
		timer_start()
		fmt.Println("Obstruction removed, closing door soon.")
	}
}

func HandleChangeInStopBtn(d bool, numFloors int) {
	// if stop button is pressed, print it
	// then un-light all button lamps
	// TODO: choose ourselves
	// TODO bare minimum set the stop light
	fmt.Println("Stop button pressed?", "%+v\n", d)
	for f := 0; f < numFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ { // initializes b to 0 (ButtonType is int)
			elevio.SetButtonLamp(b, f, false)
		}
	}
}
