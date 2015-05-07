package coordinator


import (
	"net"
	"encoding/gob"
	"../proto"
	"fmt"

	"bytes"
	"util"
)


func Handle(buf []byte,addr *net.UDPAddr, anonCoordinator *Coordinator, n int) {
	// decode the whole message
	event := &proto.Event{}
	err := gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(event)
	util.CheckErr(err)
	switch event.EventType {
	case proto.SERVER_REGISTER:
		handleServerRegister()
		break
	case proto.CLIENT_REGISTER_CONTROLLERSIDE:
		handleClientRegisterControllerSide();
		break
	case proto.CLIENT_REGISTER_SERVERSIDE:
		handleClientRegisterServerSide();
		break
	case proto.MESSAGE:
		handleMsg()
		break
	case proto.VOTE:
		handleVote()
		break
	case proto.ROUND_END:
		handleRoundEnd()
		break
	case proto.ANNOUNCEMENT:
		handleAnnouncement()
		break
	default:
		fmt.Println("[fatal] Unrecognized request...")
		break
	}
}



func handleServerRegister() {

}


func handleClientRegisterControllerSide() {
	fmt.Println("receive register")
}

func handleClientRegisterServerSide() {

}

// Handler for ANNOUNCEMENT event
// finish announcement and send start message signal to the clients
func handleAnnouncement() {
	// This event is triggered when server finishes announcement
	// distribute final reputation map to servers

}

func handleMsg() {

}

func handleVote() {

}

func handleRoundEnd() {

}

