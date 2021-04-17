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
		if e.Requests[e.Floor][btn] == true {
			e.Requests[e.Floor][btn] = false
		}
	}

	switch e.Config.ClearRequestVariant {
	case config.CV_All:
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
		break

	case config.CV_InDir:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dir {
		case elevio.MD_Up:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			if single_elev_requests.Requests_above(e) == false {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			break

		case elevio.MD_Down:
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			if single_elev_requests.Requests_below(e) == false {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			break

		case elevio.MD_Stop:
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			break
		}
		break

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
		break
	case elevator.EB_Moving:
		duration = duration + config.TRAVEL_TIME/2
		if e.Dir == elevio.MD_Up {
			e.Floor = e.Floor + 1
		}
		if e.Dir == elevio.MD_Down {
			e.Floor = e.Floor - 1
		}
		break
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
			tempelevator.Requests[order.Floor][1] = true
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
	return elevatorcostid[index] //Check this
}

func RedistributeOrders(CurrentElevStates map[string]elevator.Elevator, lostpeer string) map[string]elevator.Elevator {
	for id, elev := range CurrentElevStates {
		if lostpeer == elev.Id {
			for j := 0; j < 4; j++ {
				if elev.Requests[j][0] {
					tempbtn := elevio.ButtonEvent{j, elevio.BT_HallUp}
					communication.SendNewHallRequest(tempbtn)
					designatedElev := CurrentElevStates[id]
					designatedElev.Requests[j][0] = true
					CurrentElevStates[id] = designatedElev

				}
				if elev.Requests[j][1] {
					tempbtn := elevio.ButtonEvent{j, elevio.BT_HallDown} //? "elevator_project/elevio.ButtonEvent composite literal uses unkeyed fields"?
					communication.SendNewHallRequest(tempbtn)
					designatedElev := CurrentElevStates[id]
					designatedElev.Requests[j][0] = true
					CurrentElevStates[id] = designatedElev
				}
			}
		}
	}
	fmt.Print("Orders successfully redistributed")
	return CurrentElevStates
}
