package elevator

import (
	"elevator_project/config"
	"elevator_project/elevio"
)

type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = 0
	EB_DoorOpen                   = -1
	EB_Moving                     = 1
)

//? Kan man sette default-verdier, f.eks. Id null?
type Elevator struct {
	Id        string
	Master    bool
	Floor     int
	Dir       elevio.MotorDirection
	Behaviour ElevatorBehaviour
	Requests  [4][3]bool //  N_FLOORS, N_BUTTONS
	Config    config.Config
	// Obstruction?
	// Stop button?
}
