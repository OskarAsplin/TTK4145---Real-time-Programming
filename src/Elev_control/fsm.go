package Elev_control

import (
	"Driver"
	"time"
)

func setAllLights() {
	for floor := 0; floor < Driver.NUMFLOORS; floor++ {
		for btn := 0; btn < Driver.NUMBUTTONS; btn++ {
			if btn == int(B_Cab) {
				Driver.ElevSetButtonLight(btn, floor, elevator.Requests[floor][btn])
			} else {
				Driver.ElevSetButtonLight(btn, floor, allExtBtns[floor][btn])
			}
		}
	}
}

func updateAllExtLights(receiveAllBtnCallsCh chan [4][2]bool, allExtBtns [4][2]bool) {
	for {
		allExtBtns = <-receiveAllBtnCallsCh
		setAllLights()
	}
}

func fsm_onInitBetweenFloors() {
	Driver.ElevSetMotorDirection(int(D_Down))
	elevator.Dir = D_Down
	elevator.Behaviour = EB_Moving
	lastFloorTime = time.Now().Unix()
	for Driver.ElevGetFloorSensorSignal() == -1 {
	}
	Driver.ElevSetMotorDirection(int(D_Idle))
}

func fsm_onRequestButtonPress(btn_floor int, btn_type Button, sendBtnCallCh chan [2]int) bool {
	switch btn_type {
	case B_Cab:
		fsm_onNewActiveRequest(btn_floor, btn_type)
		return true
	case B_HallDown:
		fsm_SendNewOrderToMaster(btn_floor, btn_type, sendBtnCallCh)
		return false
	case B_HallUp:
		fsm_SendNewOrderToMaster(btn_floor, btn_type, sendBtnCallCh)
		return false
	}
	return false
}

func fsm_onNewActiveRequest(btn_floor int, btn_type Button) {
	switch elevator.Behaviour {
	case EB_DoorOpen:
		if elevator.Floor == btn_floor {
			timer_start(3000 * time.Millisecond)
		} else {
			elevator.Requests[btn_floor][btn_type] = true
		}
		break
	case EB_Moving:
		elevator.Requests[btn_floor][btn_type] = true
		break
	case EB_Idle:
		elevator.Requests[btn_floor][btn_type] = true
		elevator.Dir = requests_chooseDirection(elevator)
		if elevator.Dir == D_Idle {
			Driver.ElevSetDoorLight(true)
			elevator = requests_clearAtCurrentFloor(elevator)
			timer_start(3000 * time.Millisecond)
			elevator.Behaviour = EB_DoorOpen
		} else {
			Driver.ElevSetMotorDirection(int(elevator.Dir))
			elevator.Behaviour = EB_Moving
			lastFloorTime = time.Now().Unix()
		}
		break
	}
	setAllLights()
}

func fsm_SendNewOrderToMaster(btn_floor int, btn_type Button, sendBtnCallCh chan [2]int) {
	var newBtnCall [2]int
	newBtnCall[0] = btn_floor
	newBtnCall[1] = int(btn_type)
	sendBtnCallCh <- newBtnCall
}

func Fsm_addOrder(Order [2]int, Order_ID int64) {
	if Order_ID == elevator.Elev_ID {
		fsm_onNewActiveRequest(Order[0], Button(Order[1]))
	}
}

func fsm_onFloorArrival(newFloor int) {
	lastFloorTime = time.Now().Unix()
	elevator.Floor = newFloor
	Driver.ElevSetFloorIndicator(newFloor)
	switch elevator.Behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			Driver.ElevSetMotorDirection(int(D_Idle))
			Driver.ElevSetDoorLight(true)
			elevator = requests_clearAtCurrentFloor(elevator)
			timer_start(3000 * time.Millisecond)
			setAllLights()
			elevator.Behaviour = EB_DoorOpen
		}
		break
	default:
		break
	}
}

func fsm_onDoorTimeout() {
	switch elevator.Behaviour {
	case EB_DoorOpen:
		elevator.Dir = requests_chooseDirection(elevator)
		Driver.ElevSetMotorDirection(int(elevator.Dir))
		Driver.ElevSetDoorLight(false)
		if elevator.Dir == D_Idle {
			elevator.Behaviour = EB_Idle
		} else {
			elevator.Behaviour = EB_Moving
			lastFloorTime = time.Now().Unix()
		}
		break
	default:
		break
	}
}

func fsm_elevatorUninitialized() {
	elevator.Dir = D_Idle
	elevator.Behaviour = EB_Idle
	elevator.Floor = Driver.ElevGetFloorSensorSignal()
	elevator.Elev_ID = GetActiveTime()
	for f := 0; f < Driver.NUMFLOORS; f++ {
		for b := 0; b < Driver.NUMBUTTONS; b++ {
			elevator.Requests[f][b] = false
		}
	}
}