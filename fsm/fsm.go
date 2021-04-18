package fsm

import (
	"elevator_project/communication"
	//"elevator_project/config"
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
}

func OnInitBetweenFloors(id string) {
	ElevState.Behaviour = elevator.EB_Moving
	ElevState.Dir = elevio.MD_Down
	elevio.SetMotorDirection(ElevState.Dir)
}

func InitCurrentElevators(N_ELEVATORS int) {
	 // rename to AllElevStates?
	 CurrentElevStates = make(map[string]elevator.Elevator)	
	elevatorNames := []string{"one", "two", "three", "four", "five"}
	for i := 0; i < N_ELEVATORS; i++ {
		CurrentElevStates[elevatorNames[i]] = elevator.Elevator{Id: elevatorNames[i], Master: false, Floor: -1, Dir: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	}
	CurrentElevStates[ElevState.Id] = ElevState
}




// old implementation
// func OnInitBetweenFloors() {
// 	// outputDevice.motorDirection(D_Down);
// 	// elevator.dirn = D_Down;
// 	// elevator.behaviour = EB_Moving;
// 	ElevState.Behaviour = elevator.EB_Moving
// 	//ElevState.Dir = elevio.MD_Down //? Hvorfor måtte jeg ikke ha denne med for at det skulle funke?
// 	elevio.SetMotorDirection(elevio.MD_Down)
// 	fmt.Println("Should stop? ", single_elev_requests.ShouldStop(ElevState))
// }

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
		if call[0] == true {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
		}
		if call[1] == true {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
		}
		if s.Id == ElevState.Id {
			if call[2] == true {
				elevio.SetButtonLamp(elevio.BT_Cab, i, true)
			}
		}
	}
}

func HandleNewElevState(s elevator.Elevator) {
	UpdateLights(s)
	for i, elev := range CurrentElevStates {
		if elev.Id == s.Id {
			CurrentElevStates[i] = s
		}
	}

	if ElevState.Id == s.Id {
		ElevState = s
		fmt.Println("I MATCHED new state:")
		fmt.Println(s)
		// stuff, TODO: check this func - Seems fine?
		ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
		elevio.SetMotorDirection(ElevState.Dir)
		if ElevState.Dir != elevio.MD_Stop {
			ElevState.Behaviour = elevator.EB_Moving
		} else {
			ElevState.Behaviour = elevator.EB_Idle
			ElevState.Requests[ElevState.Floor][0] = false
			ElevState.Requests[ElevState.Floor][1] = false
			communication.SendClearedOrder(ElevState.Floor)
		}
		communication.SendStateUpdate(ElevState)
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
			communication.SendStateUpdate(ElevState) // Hvilke state updates trenger master?
			communication.SendClearedOrder(btn.Floor)
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
				elevio.SetMotorDirection(ElevState.Dir)
				ElevState.Behaviour = elevator.EB_Moving
				communication.SendStateUpdate(ElevState)
			} else { // Hall request
				communication.SendNewHallRequest(btn)
				communication.SendStateUpdate(ElevState)
				//if !ElevState.Master { // send to master
				//	communication.SendNewHallRequest(btn)
				//} //
				//else if ElevState.Master {
				// check connection to others in some way
				// then send to new hall request (or some function that does the same as new hall request receiver)
				//}
			}
		}
	case elevator.EB_DoorOpen:
		if ElevState.Floor == btn.Floor {
			timer_start()
			communication.SendClearedOrder(btn.Floor)
			fmt.Println("Extending door opening timer")
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				communication.SendStateUpdate(ElevState)
			} else { // hall request
				communication.SendNewHallRequest(btn)
				if ElevState.Master {
					communication.SendStateUpdate(ElevState)
				}
			}
		}

	case elevator.EB_Moving:
		if btn.Button == elevio.BT_Cab {
			ElevState.Requests[btn.Floor][btn.Button] = true
			elevio.SetButtonLamp(btn.Button, btn.Floor, true)
			communication.SendStateUpdate(ElevState)
		} else { // Hall request
			communication.SendNewHallRequest(btn)
			if ElevState.Master {
				communication.SendStateUpdate(ElevState)
			}
		}
	}
}

func HandleNewFloor(floor int, numFloors int) {
	fmt.Println("Arriving at floor: ", floor)

	// Set new state and send it to other elevators
	ElevState.Floor = floor
	elevio.SetFloorIndicator(ElevState.Floor)
	communication.SendStateUpdate(ElevState)

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
			for btn := elevio.ButtonType(0); btn < 3; btn++ {
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
			communication.SendClearedOrder(floor)
			communication.SendStateUpdate(ElevState)
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
