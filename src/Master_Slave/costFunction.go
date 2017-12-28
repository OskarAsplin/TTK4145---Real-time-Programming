package Master_Slave

import (
	"Elev_control"
)

func cost_function(btn_floor int, btn_type Elev_control.Button, elevs_online []Elev_control.Elevator) int {
	i_best_time := -1
	best_time := 100
	var time_to_handle int
	var floors_between int

	for i, elev := range elevs_online {
		if elev.Error == true {
			continue
		}
		floors_between = 0
		time_to_handle = 0
		if elev.Floor == btn_floor {
			if elev.Behaviour == Elev_control.EB_Idle || elev.Behaviour == Elev_control.EB_DoorOpen {
				return i
			}
		}
		switch elev.Dir {
		case Elev_control.D_Down:
			if btn_type == Elev_control.B_HallUp {
				floors_between += elev.Floor + btn_floor
			} else { //B_HallDown
				if elev.Floor <= btn_floor { 
					floors_between = 10
					break
				}
				floors_between += elev.Floor - btn_floor
			}
		case Elev_control.D_Idle:
			if elev.Floor == btn_floor {
				return i
			} else if elev.Floor > btn_floor {
				floors_between += elev.Floor - btn_floor
			} else { 
				floors_between += btn_floor - elev.Floor
			}
		case Elev_control.D_Up:
			if btn_type == Elev_control.B_HallDown {
				floors_between += 6 - elev.Floor - btn_floor
			} else {
				if elev.Floor >= btn_floor { 
					floors_between = 10
					break
				}
				floors_between += btn_floor - elev.Floor
			}
		}

		time_to_handle = calculate_time(floors_between)
		if time_to_handle < best_time {
			i_best_time = i
			best_time = time_to_handle
		}
	}

	return i_best_time
}

func calculate_time(floors_between int) int {
	time_between_floors := 1
	door_open_time := 1
	return floors_between*(time_between_floors+door_open_time) - door_open_time
}
