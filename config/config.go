package config

const N_ELEVATORS = 3
const N_FLOORS = 4
const N_BUTTONS = 3
const TRAVEL_TIME = 2
const DOOR_OPEN_TIME = 6

const MAX_ACKNOWLEDGES_TO_SEND_HallRequest = 20
const MAX_ACKNOWLEDGES_TO_SEND_StateUpdate = 10
const MAX_ACKNOWLEDGES_TO_SEND_ClearedOrder = 10

type ClearRequestVariant int

const (
	CV_All = iota
	CV_InDir
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	doorOpenDuration_s  int
}
