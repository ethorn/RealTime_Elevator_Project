package elevator

import (
	"elevator_project/elevio"
)

type ElevatorBehaviour int
const (
	EB_Idle 	  ElevatorBehaviour = 0
    EB_DoorOpen				  		= -1
    EB_Moving				  		= 1
)

type Elevator struct {
	Id string
	Master bool
    Floor int
    Dir elevio.MotorDirection
    Behaviour ElevatorBehaviour
	Requests[4][3] bool //  N_FLOORS, N_BUTTONS
	// Obstruction?
	// Stop button?
}