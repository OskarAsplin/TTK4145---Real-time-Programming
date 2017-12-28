package Elev_control

import (
	"Driver"
)

type Direction int

const (
	D_Down Direction = -1 + iota
	D_Idle
	D_Up
)

type Button int

const (
	B_HallDown Button = 0 + iota
	B_HallUp
	B_Cab
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = 0 + iota
	EB_DoorOpen
	EB_Moving
)

func requests_above(e Elevator) bool {
	for f := e.Floor + 1; f < Driver.NUMFLOORS; f++ {
		for btn := 0; btn < Driver.NUMBUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < Driver.NUMBUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_chooseDirection(e Elevator) Direction {
	switch e.Dir {
	case D_Up:
		if requests_above(e) {
			return D_Up
		} else if requests_below(e) {
			return D_Down
		} else {
			return D_Idle
		}
	case D_Down:
		if requests_below(e) {
			return D_Down
		} else if requests_above(e) {
			return D_Up
		} else {
			return D_Idle
		}
	case D_Idle: 
		if requests_above(e) {
			return D_Up
		} else if requests_below(e) {
			return D_Down
		} else {
			return D_Idle
		}
	default:
		return D_Idle
	}
}

func requests_shouldStop(e Elevator) bool {
	switch e.Dir {
	case D_Down:
		return e.Requests[e.Floor][B_HallDown] || e.Requests[e.Floor][B_Cab] || !requests_below(e)
	case D_Up:
		return e.Requests[e.Floor][B_HallUp] || e.Requests[e.Floor][B_Cab] || !requests_above(e)
	case D_Idle:
		return true
	default:
		return true
	}
}

func requests_clearAtCurrentFloor(e Elevator) Elevator {
	e.Requests[e.Floor][B_Cab] = false
	e.Requests[e.Floor][B_HallUp] = false
	e.Requests[e.Floor][B_HallDown] = false
	return e
}
