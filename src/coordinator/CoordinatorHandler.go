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


// handle server register request
func handleServerRegister() {
	fmt.Println("[debug] Receive the registration info from server " + srcAddr.String());
	// send reply to the new server
	lastServer := anonCoordinator.GetLastServer()
	pm1 := map[string]interface{}{
		"reply": true,
		"prev_server": lastServer.String(),
	}
	event1 := &proto.Event{proto.SERVER_REGISTER_REPLY,pm1}
	util.Send(anonCoordinator.Socket,srcAddr,util.Encode(event1))

	// update next hop for previous server

	if (lastServer != nil) {
		pm2 := map[string]interface{}{
			"reply": true,
			"next_hop": srcAddr.String(),
		}
		event2 := &proto.Event{proto.UPDATE_NEXT_HOP, pm2}
		util.Send(anonCoordinator.Socket, srcAddr, util.Encode(event2))
	}
	anonCoordinator.AddServer(srcAddr);
}

// Handler for REGISTER event
// send the register request to server to do encryption
func handleClientRegisterControllerSide(params map[string]interface{}) {
	// get client's public key
	publicKey := anonCoordinator.Suite.Point()
	publicKey.UnmarshalBinary(params["public_key"].([]byte))
	anonCoordinator.AddClient(publicKey,srcAddr)

	// send register info to the first server
	firstServer := anonCoordinator.GetFirstServer()
	pm := map[string]interface{}{
		"public_key": params["public_key"],
		"addr": srcAddr.String(),
	}
	event := &proto.Event{proto.CLIENT_REGISTER_SERVERSIDE,pm}
	util.Send(anonCoordinator.Socket,firstServer,util.Encode(event))

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

