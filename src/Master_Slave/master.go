package Master_Slave

import (
	"Elev_control"
	"Network"
	"fmt"
)

var isMaster bool

func Run_MasterSlave_Logic(localStatusCh chan Elev_control.Elevator, sendBtnCallCh chan [2]int, receiveAllBtnCallsCh, setLights_setExtBtnsCh chan [4][2]bool, errorCh chan int) {
	msgToNetwork := make(chan Network.UdpMessage, 10)
	msgFromNetwork := make(chan Network.UdpMessage, 10)
	updateElevsCh := make(chan Elev_control.Elevator, 10)
	isMasterCh := make(chan bool)
	sendOrderCh := make(chan Network.UdpMessage, 10)
	receiveBtnCallCh := make(chan [2]int, 10)
	handleOrderAgainCh := make(chan [2]int, 10)

	isMaster = true
	wasMaster := false
	Network.MH_UpdateMasterStatus(isMaster)
	master_elev := <-localStatusCh
	update_Elevators_online(master_elev)

	Network.Init_udp(msgToNetwork, msgFromNetwork, isMasterCh)
	go Network.MH_HandleIncomingMsg(msgFromNetwork, updateElevsCh, receiveBtnCallCh, receiveAllBtnCallsCh)
	go Network.MH_HandleOutgoingMsg(msgToNetwork, sendOrderCh, updateElevsCh, localStatusCh, sendBtnCallCh, receiveBtnCallCh)
	var call [2]int
	for {
		select {
		case call = <-receiveBtnCallCh:
			if update_btnCalls(call) {
				temp_Elevators_online := getElevators_Online()
				index_elev := cost_function(call[0], Elev_control.Button(call[1]), temp_Elevators_online)
				if index_elev == -1 {
					errorCh <- Elev_control.ERR_NO_ELEVS_OPERABLE
					break
				}
				elev_ID := temp_Elevators_online[index_elev].Elev_ID
				Network.MH_send_new_order(elev_ID, call, sendOrderCh)
			}
		case call = <-handleOrderAgainCh:
			temp_Elevators_online := getElevators_Online()
			index_elev := cost_function(call[0], Elev_control.Button(call[1]), temp_Elevators_online)
			if index_elev == -1 {
				errorCh <- Elev_control.ERR_NO_ELEVS_OPERABLE
				break
			}
			elev_ID := temp_Elevators_online[index_elev].Elev_ID
			fmt.Printf("%+v", elev_ID)
			Network.MH_send_new_order(elev_ID, call, sendOrderCh)
		case allCalls := <-receiveAllBtnCallsCh:
			setAll_btn_calls(allCalls)
			setLights_setExtBtnsCh <- allCalls
		case elev := <-updateElevsCh:
			update_All_elevs(elev)
			check_elevsIdleAtFloor()
		case isMaster = <-isMasterCh:
			if isMaster {
				go setNewTimeStampsOnActiveOrders()
				fmt.Println("                        				Er master")
				wasMaster = true
			} else {
				fmt.Println("                        				Er slave")
				if wasMaster{
					Network.MH_broadcast_all_btn_calls(get_All_btn_calls(), sendOrderCh)		
				}
				wasMaster = false
			}
			delete_All_elevs()
			Network.MH_UpdateMasterStatus(isMaster)
		}
		if isMaster {
			Network.MH_broadcast_all_btn_calls(get_All_btn_calls(), sendOrderCh)
			setLights_setExtBtnsCh <- get_All_btn_calls()
			checkTimeStamps(handleOrderAgainCh)
		}
	}
}