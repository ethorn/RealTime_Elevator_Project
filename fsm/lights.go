package fsm

import (
	"elevator_project/config"
	"elevator_project/elevio"
)

func UpdateAllLights() {
	var HasHallDownRequests map[int]bool = make(map[int]bool)
	var HasHallUpRequests map[int]bool = make(map[int]bool)
	for f := 0; f < config.N_FLOORS; f++ {
		HasHallDownRequests[f] = false
		HasHallUpRequests[f] = false
	}
	for id := range CurrentElevStates {
		for floor := range CurrentElevStates[id].Requests {
			if CurrentElevStates[id].Requests[floor][elevio.BT_HallDown] {
				HasHallDownRequests[floor] = true
			}
			if CurrentElevStates[id].Requests[floor][elevio.BT_HallUp] {
				HasHallUpRequests[floor] = true
			}
		}
	}
	
	for f := 0; f < config.N_FLOORS; f++ {
		elevio.SetButtonLamp(elevio.BT_HallDown, f, HasHallDownRequests[f])
		elevio.SetButtonLamp(elevio.BT_HallUp, f, HasHallUpRequests[f])
		elevio.SetButtonLamp(elevio.BT_Cab, f, ElevState.Requests[f][elevio.BT_Cab])
	}
}