package config

const N_ELEVATORS = 3
const N_FLOORS = 4
const N_BUTTONS = 3
const TRAVEL_TIME = 5
const DOOR_OPEN_TIME = 5

type ClearRequestVariant int

const (
	CV_All = iota
	CV_InDir
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	doorOpenDuration_s  int
}
