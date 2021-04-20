package order_logic

import (
	"elevator_project/communication"
	"elevator_project/config"
	"elevator_project/elevator"
	"elevator_project/elevio"
	"elevator_project/single_elev_requests"
	"fmt"
)

func Requests_clearAtCurrentFloor(e_old elevator.Elevator) elevator.Elevator {
	e := e_old
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			e.Requests[e.Floor][btn] = false
		}
	}

	switch e.Config.ClearRequestVariant {
	case config.CV_All:
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}

	case config.CV_InDir:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dir {
		case elevio.MD_Up:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			if single_elev_requests.Requests_above(e) {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}

		case elevio.MD_Down:
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			if single_elev_requests.Requests_below(e) {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}

		case elevio.MD_Stop:
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}

	default:
		break
	}

	return e
}

func TimeToIdle(e elevator.Elevator) int {
	duration := 0

	switch e.Behaviour {
	case elevator.EB_Idle:
		e.Dir = single_elev_requests.ChooseDirection(e)
		if e.Dir == elevio.MD_Stop {
			return duration
		}
	case elevator.EB_Moving:
		duration = duration + config.TRAVEL_TIME/2
		if e.Dir == elevio.MD_Up {
			e.Floor = e.Floor + 1
		}
		if e.Dir == elevio.MD_Down {
			e.Floor = e.Floor - 1
		}
	case elevator.EB_DoorOpen:
		duration = config.DOOR_OPEN_TIME/2 - duration
	}

	for {
		if single_elev_requests.ShouldStop(e) {
			e = Requests_clearAtCurrentFloor(e)
			duration = duration + config.DOOR_OPEN_TIME
			e.Dir = single_elev_requests.ChooseDirection(e)
			if e.Dir == elevio.MD_Stop {
				return duration
			}
		}
		if e.Dir == elevio.MD_Up {
			e.Floor = e.Floor + 1
		}
		if e.Dir == elevio.MD_Down {
			e.Floor = e.Floor - 1
		}
		duration = duration + config.TRAVEL_TIME
	}
}

func DesignateOrder(CurrentElevStates map[string]elevator.Elevator, order elevio.ButtonEvent, unservicablePeers []string) string {
	var elevatorcost []int
	var elevatorcostid []string
	index := 0
	runCost := true
	for _, elevator := range CurrentElevStates {
		if len(unservicablePeers) > 0 {
			for _, blocked := range unservicablePeers {
				if blocked == elevator.Id {
					runCost = false
				}
			}
		}

		if runCost == true {
			tempelevator := elevator
			tempelevator.Requests[order.Floor][order.Button] = true
			elevatorcost = append(elevatorcost, TimeToIdle(tempelevator))
			elevatorcostid = append(elevatorcostid, elevator.Id)
		}
		runCost = true
	}

	min := elevatorcost[0]
	for i, cost := range elevatorcost {
		if cost < min {
			min = cost
			index = i

		}
	}
	fmt.Println("New request assigned to elevator", elevatorcostid[index])
	return elevatorcostid[index]
}

func RedistributeOrders(CurrentElevStates map[string]elevator.Elevator, lostpeer string, senderId string) map[string]elevator.Elevator {
	for id, elev := range CurrentElevStates {
		if lostpeer == elev.Id {
			for j := 0; j < config.N_FLOORS; j++ {
				if elev.Requests[j][0] {
					tempbtn := elevio.ButtonEvent{j, elevio.BT_HallUp}
					communication.SendNewHallRequest(tempbtn, senderId)
					removedElev := CurrentElevStates[id]
					removedElev.Requests[j][0] = false
					CurrentElevStates[id] = removedElev
				}
				if elev.Requests[j][1] {
					tempbtn := elevio.ButtonEvent{j, elevio.BT_HallDown}
					communication.SendNewHallRequest(tempbtn, senderId)
					removedElev := CurrentElevStates[id]
					removedElev.Requests[j][1] = false
					CurrentElevStates[id] = removedElev
				}
			}
		}
	}
	fmt.Println("Orders successfully redistributed")
	return CurrentElevStates
}
