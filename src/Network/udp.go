package Network

import (
	"Elev_control"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

const(
	MSGsize = 512
	masterPort = 47838
	slavePort = 47839
	alive_port = "47837"
)

type UdpMessage struct {
	Data      Elev_control.Elevator
	Length    int 
	Order_ID  int64
	Order     [2]int
	Btn_calls [4][2]bool
}

func Init_udp(msgToNetwork, msgFromNetwork chan UdpMessage, isMasterCh chan bool) {
	my_IP := GetLocalIP()
	fmt.Println("Lokal ip_addresse: \n", my_IP)
	peerListLocalCh := make(chan []string)
	msgToNetwork_master := make(chan UdpMessage, 100)
	msgToNetwork_slave := make(chan UdpMessage, 100)
	msgFromNetwork_master := make(chan UdpMessage, 100)
	msgFromNetwork_slave := make(chan UdpMessage, 100)
	go transmitMsg(masterPort, msgToNetwork_master)
	go transmitMsg(slavePort, msgToNetwork_slave)
	go receiveMsg(masterPort, msgFromNetwork_master)
	go receiveMsg(slavePort, msgFromNetwork_slave)
	go udpSendAlive(alive_port)
	go udpRecvAlive(alive_port, peerListLocalCh)

	go func() {
		isMaster := false
		for {
			select {
			case msg := <-msgFromNetwork_slave:
				if isMaster {
					msgFromNetwork <- msg
				}
			case msg := <-msgFromNetwork_master:
				if !isMaster {
					msgFromNetwork <- msg
				}
			case msg := <-msgToNetwork:
				if isMaster {
					msgToNetwork_master <- msg
				} else {
					msgToNetwork_slave <- msg
				}
			case new_peer_list := <-peerListLocalCh:
				fmt.Println(new_peer_list)
				highest_IP := my_IP
				for _, IP := range new_peer_list {
					if highest_IP < IP {
						highest_IP = IP
					}
				}
				if my_IP == highest_IP {
					isMaster = true
					isMasterCh <- true
				} else {
					isMaster = false
					isMasterCh <- false
				}
			}
		}
	}()
}

func receiveMsg(port int, messageFromNetwork chan UdpMessage) {
	bAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(port))
	broadcastConn, err := net.ListenUDP("udp4", bAddr)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_receive_server: \nClosing connection.", r)
			broadcastConn.Close()
		}
	}()

	if err != nil {
		broadcastConn.Close()
	}

	for {
		buf := make([]byte, MSGsize)
		n, _, err := broadcastConn.ReadFromUDP(buf)
		if err != nil {
			broadcastConn.Close()
		}
		buf = buf[:n]
		var msg UdpMessage
		DecodeMessage(&msg, buf)
		messageFromNetwork <- msg
	}
}

func transmitMsg(port int, messageToNetwork chan UdpMessage) {
	bAddr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(port))
	broadcastConn, _ := net.DialUDP("udp4", nil, bAddr)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udpTransmitServer: %s \n Closing connection.", r)
			broadcastConn.Close()
		}
	}()
	for {
		msg := <-messageToNetwork
		broadcastConn.Write(EncodeMessage(msg))
	}
}

func EncodeMessage(msg UdpMessage) []byte {
	returnMessage, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Error: problems encoding UdpMessage struct.\n")
		panic(err)
	}
	return returnMessage
}

func DecodeMessage(msg *UdpMessage, arr []byte) {
	json.Unmarshal(arr, msg)
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}