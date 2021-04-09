package single_elev_requests

import (
	"elevator_project/config"
	"elevator_project/elevator"
	"elevator_project/elevio"
)

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dir {
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !Requests_above(e)
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !Requests_below(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func ChooseDirection(e elevator.Elevator) elevio.MotorDirection {
	// Need to continue in the direction it was going, if there are any orders in that direction left

	switch e.Dir {
	case elevio.MD_Up:
		if Requests_above(e) {
			return elevio.MD_Up
		} else if Requests_below(e) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	case elevio.MD_Down:
		if Requests_below(e) {
			return elevio.MD_Down
		} else if Requests_above(e) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	case elevio.MD_Stop:
		if Requests_below(e) {
			return elevio.MD_Down
		} else if Requests_above(e) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	default:
		return elevio.MD_Stop
	}
}

func Requests_above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < config.N_FLOORS; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}
