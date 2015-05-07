package coordinator


import (
	"net"
	"encoding/gob"
	"../proto"
	"fmt"

	"bytes"
	"util"
)

var anonCoordinator *Coordinator
var srcAddr *net.UDPAddr
func Handle(buf []byte,addr *net.UDPAddr, tmpCoordinator *Coordinator, n int) {
	// decode the whole message
	anonCoordinator = tmpCoordinator
	srcAddr = addr

	event := &proto.Event{}
	err := gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(event)
	util.CheckErr(err)
	switch event.EventType {
	case proto.SERVER_REGISTER:
		handleServerRegister()
		break
	case proto.CLIENT_REGISTER_CONTROLLERSIDE:
		handleClientRegisterControllerSide(event.Params,);
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



// Handler for REGISTER event
// send the register request to server to do encryption
func handleClientRegisterControllerSide(params map[string]interface{}) {
	// get client's public key
	publicKey := anonCoordinator.Suite.Point()
	publicKey.UnmarshalBinary(params["public_key"].([]byte))
	anonCoordinator.AddClient(publicKey,srcAddr)

	// send register info to the first server
//	firstServer := anonCoordinator.GetFirstServer()
	pm := map[string]interface{}{
		"public_key": params["public_key"],
		"addr": srcAddr.String(),
	}
	event := &proto.Event{proto.CLIENT_REGISTER_SERVERSIDE,pm}
	util.Send(anonCoordinator.Socket,srcAddr,util.Encode(event))

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

