package Master_Slave

import (
	"Elev_control"
	"time"
)

const (
	NUMBUTTONS  int = 3
	NUMFLOORS   int = 4
)

var elevators_online []Elev_control.Elevator
var all_btn_calls [NUMFLOORS][NUMBUTTONS-1]bool
var btn_calls_timeStamp [NUMFLOORS][NUMBUTTONS - 1]int64

func update_Elevators_online(curr_elev Elev_control.Elevator) {
	for i, elev := range elevators_online {
		if elev.Elev_ID == curr_elev.Elev_ID {
			elevators_online[i] = curr_elev
			return
		}
	}
	elevators_online = append(elevators_online, curr_elev)
}

func delete_All_elevs() {
	elevators_online = elevators_online[:0]
}

func update_All_elevs(elev Elev_control.Elevator) {
	update_Elevators_online(elev)
}

func update_btnCalls(newCall [2]int) bool {
	if all_btn_calls[newCall[0]][newCall[1]] {
		return false
	}
	all_btn_calls[newCall[0]][newCall[1]] = true
	setTimeStamp(newCall[0], newCall[1])
	return true
}

func getElevators_Online() []Elev_control.Elevator {
	return elevators_online
}

func setElevators_Online(elevs_online []Elev_control.Elevator) {
	elevators_online = elevs_online
}

func get_All_btn_calls() [4][2]bool {
	return all_btn_calls
}

func setAll_btn_calls(btn_calls [4][2]bool) {
	all_btn_calls = btn_calls
}

func setTimeStamp(btn_floor int, btn_type int) {
	btn_calls_timeStamp[btn_floor][btn_type] = time.Now().Unix()
}

func setNewTimeStampsOnActiveOrders() {
	for floor, btnsAtFloor := range all_btn_calls {
		for btn, call := range btnsAtFloor {
			if call {
				setTimeStamp(floor, btn)
			}
		}
	}
}

func checkTimeStamps(handleOrderAgainCh chan [2]int) {
	var errorTime int64
	var timeNow int64
	errorTime = 10
	var order [2]int
	if isMaster {
		timeNow = time.Now().Unix()
		for i, k := range btn_calls_timeStamp {
			for j, timeStamp := range k {
				if timeStamp != 0 {
					if timeNow-timeStamp > errorTime {
						order[0] = i
						order[1] = j
						handleOrderAgainCh <- order
						setTimeStamp(i, j)
					}
				}
			}
		}
	}
}

func check_elevsIdleAtFloor() {
	if isMaster {
		for _, elev := range elevators_online {
			if elev.Behaviour == Elev_control.EB_Idle || elev.Behaviour == Elev_control.EB_DoorOpen {
				all_btn_calls[elev.Floor][Elev_control.B_HallDown] = false
				btn_calls_timeStamp[elev.Floor][Elev_control.B_HallDown] = 0
				all_btn_calls[elev.Floor][Elev_control.B_HallUp] = false
				btn_calls_timeStamp[elev.Floor][Elev_control.B_HallUp] = 0
			}
		}
	}
}