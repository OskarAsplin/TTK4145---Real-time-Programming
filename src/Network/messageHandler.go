package Network

import (
	"Elev_control"
)

var isMaster bool

const (
	statusMsg      int64 = 0
	btnCallMsg     int64 = 1
	allBtnCallsMsg int64 = 2
)

func MH_HandleIncomingMsg(msgFromNetwork chan UdpMessage, updateElevsCh chan Elev_control.Elevator, receiveBtnCallCh chan [2]int, receiveAllBtnCallsCh chan [4][2]bool) {
	var msg UdpMessage
	for {
		msg = <-msgFromNetwork
		switch msg.Order_ID {
		case statusMsg:
			updateElevsCh <- msg.Data
			break
		case btnCallMsg:
			receiveBtnCallCh <- msg.Order
		case allBtnCallsMsg:
			receiveAllBtnCallsCh <- msg.Btn_calls
		default:
			Elev_control.Fsm_addOrder(msg.Order, msg.Order_ID)

		}
	}
}

func MH_HandleOutgoingMsg(msgToNetwork, sendOrderCh chan UdpMessage, updateElevsCh, localStatusCh chan Elev_control.Elevator, sendBtnCallCh, receiveBtnCallCh chan [2]int) {
	var elev Elev_control.Elevator
	var msg UdpMessage
	for {
		if isMaster {
			select {
			case elev = <-localStatusCh:
				updateElevsCh <- elev
			case msg = <-sendOrderCh:
				msgToNetwork <- msg
			case btn_call := <-sendBtnCallCh:
				receiveBtnCallCh <- btn_call
			}
		} else {
			select {
			case elev = <-localStatusCh:
				msg.Order_ID = 0
				msg.Data = elev
				msgToNetwork <- msg
			case new_call := <-sendBtnCallCh:
				msg.Order_ID = 1
				msg.Order = new_call
				msgToNetwork <- msg
			}
		}
	}
}

func MH_UpdateMasterStatus(isMasterFrom_Master_Slave bool) {
	isMaster = isMasterFrom_Master_Slave
}

func MH_send_new_order(to_elev int64, order [2]int, sendOrderCh chan UdpMessage) {
	var msg UdpMessage
	msg.Order_ID = to_elev
	msg.Order = order
	Elev_control.Fsm_addOrder(order, to_elev)
	sendOrderCh <- msg
}

func MH_broadcast_all_btn_calls(all_btn_calls [4][2]bool, sendOrderCh chan UdpMessage) {
	var msg UdpMessage
	msg.Order_ID = 2
	msg.Btn_calls = all_btn_calls
	sendOrderCh <- msg
}
