package Elev_control

import (
	"Driver"
	"fmt"
	"time"
)

type Elevator struct {
	Floor     int
	Dir       Direction
	Requests  [Driver.NUMFLOORS][Driver.NUMBUTTONS]bool
	Behaviour ElevatorBehaviour
	Elev_ID   int64
	Error     bool
}

var (
	elevator      Elevator
	allExtBtns    [Driver.NUMFLOORS][Driver.NUMBUTTONS - 1]bool
	lastFloorTime int64
)

const (
	ERR_NO_ERROR = 0 + iota
	ERR_MOTORSTOP
	ERR_NO_ELEVS_OPERABLE
)


func Run_Elevator(localStatusCh chan Elevator, sendBtnCallCh chan [2]int, receiveAllBtnCallsCh chan [Driver.NUMFLOORS][Driver.NUMBUTTONS - 1]bool, setLights_setExtBtnsCh chan [4][2]bool, errorCh chan int) {
	if Driver.ElevGetFloorSensorSignal() == -1 {
		fsm_onInitBetweenFloors()
	}
	fsm_elevatorUninitialized()
	fmt.Println("Elev ID: ", elevator.Elev_ID)

	running := true
	var prev_button [Driver.NUMFLOORS][Driver.NUMBUTTONS]int
	var prev_floor int
	prev_floor = Driver.ElevGetFloorSensorSignal()
	go Update_ExtBtnCallsInElevControl(setLights_setExtBtnsCh) 
	go checkElevMoving(errorCh)
	update_status_count := 0
	update_lights_count := 0
	localStatusCh <- elevator
	for running {
		for f := 0; f < Driver.NUMFLOORS; f++ {
			for b := 0; b < Driver.NUMBUTTONS; b++ {
				v := Driver.ElevGetButtonSignal(b, f)
				if v&int(v) != prev_button[f][b] {
					if fsm_onRequestButtonPress(f, Button(b), sendBtnCallCh) { //Hvis true er det innvendig knappetrykk
						temp_Behaviour := elevator.Behaviour
						fsm_onNewActiveRequest(f, Button(b))
						if temp_Behaviour != elevator.Behaviour {
							localStatusCh <- elevator
						}
					} else {
						fsm_SendNewOrderToMaster(f, Button(b), sendBtnCallCh)
					}
				}
				prev_button[f][b] = v
			}
		}
		f := Driver.ElevGetFloorSensorSignal()
		if f != -1 {
			if f != prev_floor {
				fsm_onFloorArrival(f)
				localStatusCh <- elevator
			}
		}
		prev_floor = f
		if timer_timedOut() {
			fsm_onDoorTimeout()
			timer_stop()
			localStatusCh <- elevator
		}
		if update_status_count == 50 {
			localStatusCh <- elevator
			update_status_count = 0
		}
		if update_lights_count == 10 {
			setLights_setExtBtnsCh <- allExtBtns
			update_lights_count = 0
		}
		time.Sleep(25 * time.Millisecond) 
		update_status_count += 1
		update_lights_count += 1
	}
}

func checkElevMoving(errorCh chan int) {
	var errorTime int64
	var timeNow int64
	errorTime = 5
	var err int
	for {
		time.Sleep(3 * time.Second)
		if elevator.Behaviour == EB_Moving {
			timeNow = time.Now().Unix()
			if timeNow-lastFloorTime > errorTime {
				err = ERR_MOTORSTOP
				errorCh <- err
				elevator.Error = true
			}
		} else {
			err = ERR_NO_ERROR
			elevator.Error = false
			errorCh <- err
		}
	}
}

func Update_ExtBtnCallsInElevControl(setLights_setExtBtnsCh chan [4][2]bool) {
	for {
		allExtBtns = <-setLights_setExtBtnsCh
		setAllLights()
	}
}