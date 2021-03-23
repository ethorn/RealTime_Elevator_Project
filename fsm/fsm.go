package fsm

import (
	"elevator_project/elevio"
	"fmt"
	"elevator_project/communication"
	"elevator_project/elevator"
	"elevator_project/single_elev_requests"
	"time"
)
// IMPORTANT: ALL STATE CHANGES HAPPEN IN THIS FILE

var d elevio.MotorDirection
var ElevState elevator.Elevator
var timerEndTime time.Time
var timerActive bool

func init() {
    initElevator()
}

func initElevator() {
	ElevState = elevator.Elevator{Id: "null", Master: false, Floor: -1, Dir: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
}

func onInitBetweenFloors() {
	// outputDevice.motorDirection(D_Down);
    // elevator.dirn = D_Down;
    // elevator.behaviour = EB_Moving;
}

func HandleNewElevState(s elevator.Elevator) {
	if ElevState.Id == s.Id {
		ElevState = s
		// stuff, TODO: check this func
		ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
		elevio.SetMotorDirection(ElevState.Dir)
		if ElevState.Dir != elevio.MD_Stop {
			ElevState.Behaviour = elevator.EB_Moving;
		} else {
			ElevState.Behaviour = elevator.EB_Idle;
		}
	} else {
		communication.StatesUpdateTx<-s
	}
}

// Handle functions
func HandleDoorTimeOut(e bool) {
	switch ElevState.Behaviour {
	case elevator.EB_DoorOpen:
		ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
		elevio.SetDoorOpenLamp(false)
		fmt.Println("closing door")
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
	fmt.Printf("%+v\n", btn) // debuggin: prints out the button event

	switch ElevState.Behaviour {
	case elevator.EB_Idle:
		if ElevState.Floor == btn.Floor {
			elevio.SetDoorOpenLamp(true)
			timer_start()
			ElevState.Behaviour = elevator.EB_DoorOpen
			communication.SendStateUpdate(ElevState) // Hvilke state updates trenger master?
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				ElevState.Dir = single_elev_requests.ChooseDirection(ElevState)
				elevio.SetMotorDirection(ElevState.Dir)
				ElevState.Behaviour = elevator.EB_Moving;
				communication.SendStateUpdate(ElevState)
			} else { // Hall request
				if !ElevState.Master { // send to master
					communication.SendNewHallRequest(btn)
				} else if ElevState.Master {
					// TODO
					// check connection to others in some way
					// then send to new hall request (or some function that does the same as new hall request receiver)
				}
			}
		}
	case elevator.EB_DoorOpen:
		if ElevState.Floor == btn.Floor {
			timer_start()
		} else {
			if btn.Button == elevio.BT_Cab {
				ElevState.Requests[btn.Floor][btn.Button] = true
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				communication.SendStateUpdate(ElevState)
			} else { // hall request
				if !ElevState.Master { // send to master
					communication.SendNewHallRequest(btn)
				} else if ElevState.Master {
					// TODO
					// check connection to others in some way
					// then send to new hall request (or some function that does the same as new hall request receiver)
				}
			}
		}
		
	case elevator.EB_Moving:
		if btn.Button == elevio.BT_Cab {
			ElevState.Requests[btn.Floor][btn.Button] = true
			elevio.SetButtonLamp(btn.Button, btn.Floor, true)
			communication.SendStateUpdate(ElevState)
		} else { // Hall request
			if !ElevState.Master { // send to master
				communication.SendNewHallRequest(btn)
			} else if ElevState.Master {
				// TODO
				// check connection to others in some way
				// then send to new hall request (or some function that does the same as new hall request receiver)
			}
		}
	}

	// fmt.Println("New state:")
	// elevator_print(elevator) TODO
}

func HandleNewFloor(floor int, numFloors int) {
	// If new floor (not same floor, or not in between), print floor
	// And change direction if in bottom or top

	// TODO

	fmt.Printf("%+v\n", floor)

	ElevState.Floor = floor
	elevio.SetFloorIndicator(ElevState.Floor)
	communication.SendStateUpdate(ElevState)

	switch ElevState.Behaviour {
	case elevator.EB_Moving:
		if single_elev_requests.ShouldStop(ElevState) {
			for i := elevio.ButtonType(0); i < 3; i++ {
				if ElevState.Requests[floor][i] {
					// Stop
					ElevState.Dir = elevio.MD_Stop
					elevio.SetMotorDirection(ElevState.Dir)
					
					elevio.SetDoorOpenLamp(true)
					timer_start()
					ElevState.Behaviour = elevator.EB_DoorOpen;
					
					if i == elevio.BT_HallUp {
						fmt.Println("entering: Hall up please")
					} else if i == elevio.BT_HallDown {
						fmt.Println("entering: Hall down please")
					} else if i == elevio.BT_Cab {
						fmt.Println("leaving: cab")
					}
					ElevState.Requests[floor][i] = false
					elevio.SetButtonLamp(i, floor, false)
					communication.SendStateUpdate(ElevState)
				}
			}
		}
	}
}

func HandleChangeInObstacle(obstruction bool) {
	// TODO
		// Should keep door open if obstructed (renew timer?)
	fmt.Printf("%+v\n", obstruction)
	// if obstruction {
	// 	elevio.SetMotorDirection(elevio.MD_Stop)
	// 	timer_start()
	// } else {
	// 	elevio.SetMotorDirection(d)
	// }
}

func HandleChangeInStopBtn(d bool, numFloors int) {
	// if stop button is pressed, print it
	// then un-light all button lamps
	// TODO: choose ourselves
	fmt.Printf("%+v\n", d)
	for f := 0; f < numFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ { // initializes b to 0 (ButtonType is int)
			elevio.SetButtonLamp(b, f, false)
		}
	}
}