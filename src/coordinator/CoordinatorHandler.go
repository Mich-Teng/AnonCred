package coordinator


import (
	"net"
	"encoding/gob"
	"../proto"
	"fmt"

	"bytes"
	"util"
	"strings"
	"strconv"
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
		handleMsg(event.Params)
		break
	case proto.VOTE:
		handleVote()
		break
	case proto.ROUND_END:
		handleRoundEnd()
		break
	case proto.ANNOUNCEMENT:
		handleAnnouncement(event.Params)
		break
	default:
		fmt.Println("[fatal] Unrecognized request...")
		break
	}
}


// Handler for ANNOUNCEMENT event
// finish announcement and send start message signal to the clients
func handleAnnouncement(params map[string]interface{}) {
	// This event is triggered when server finishes announcement
	// distribute final reputation map to servers
	var g = anonCoordinator.Suite.Point()
	byteG := params["g"].([]byte)
	err := g.UnmarshalBinary(byteG)
	util.CheckErr(err)
	// distribute g and hash table of ids to user
	pm := map[string]interface{}{
		"g": params["g"].([]byte),
	}
	event := &proto.Event{proto.ANNOUNCEMENT,pm}
	for _,val := range anonCoordinator.Clients {
		util.Send(anonCoordinator.Socket,val,util.Encode(event))
	}

	// set controller's new g
	anonCoordinator.G = g
	anonCoordinator.Status = MESSAGE
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

// handle client register successful event
func handleClientRegisterServerSide(params map[string]interface{}) {
	// get public key from params (it's one-time nym actually)
	var publicKey = anonCoordinator.Suite.Point()
	bytePublicKey := params["public_key"].([]byte)
	publicKey.UnmarshalBinary(bytePublicKey)

	var addrStr = params["addr"].(string)
	addr,err := net.ResolveUDPAddr("udp",addrStr)
	util.CheckErr(err)
	pm := map[string]interface{}{}
	event := &proto.Event{proto.CLIENT_REGISTER_SERVERSIDE,pm}
	util.Send(anonCoordinator.Socket,addr,util.Encode(event))

	// instead of sending new client to server, we will send it when finishing this round. Currently we just add it into buffer
	anonCoordinator.NewClientsBuffer = append(anonCoordinator.NewClientsBuffer,publicKey)
}

// verify the msg and broadcast to clients
func handleMsg(params map[string]interface{}) {
	// get info from the request
	text := params["text"].(string)
	byteSig := params["signature"].([]byte)
	nym := anonCoordinator.Suite.Point()
	byteNym := params["nym"].([]byte)
	err := nym.UnmarshalBinary(byteNym)
	util.CheckErr(err)

	fmt.Println("[debug] Receiving msg from " + srcAddr.String() + ": " + text)
	// verify the identification of the client

	byteText := []byte(text)
	err = util.ElGamalVerify(anonCoordinator.Suite,byteText,nym,byteSig)
	if err != nil {
		fmt.Print("[note]** Fails to verify the message...")
		return
	}
	// add msg log
	msgID := anonCoordinator.AddMsgLog(nym)

	// generate msg to clients
	pm := map[string]interface{}{
		"text" : text,
		"nym" : params["nym"].([]byte),
		"rep" : anonCoordinator.GetReputation(nym),
		"msgID" : msgID,
	}
	event := &proto.Event{proto.MESSAGE,pm}

	// send to all the clients
	for _,val := range anonCoordinator.Clients {
		util.Send(anonCoordinator.Socket,val,util.Encode(event))
	}

}

// verify the vote and reply to client
func handleVote(params map[string]interface{}) {
	// get info from the request
	text := params["text"].(string)
	byteSig := params["signature"].([]byte)
	nym := anonCoordinator.Suite.Point()
	byteNym := params["nym"].([]byte)
	err := nym.UnmarshalBinary(byteNym)
	util.CheckErr(err)

	fmt.Println("[debug] Receiving vote from " + srcAddr.String() + ": " + text)
	// verify the identification of the client

	byteText := []byte(text)
	err = util.ElGamalVerify(anonCoordinator.Suite,byteText,nym,byteSig)
	var pm map[string]interface{}
	if err != nil {
		fmt.Print("[note]** Fails to verify the vote...")
		pm = map[string]interface{}{
			"reply" : false,
		}
		return
	}else {
		// avoid duplicate vote
		// todo

		// get msg id and vote
		commands := strings.Split(text,";")
		// modify the reputation
		msgID, _ := strconv.Atoi(commands[0])
		vote, _ := strconv.Atoi(commands[1])
		targetNym := anonCoordinator.MsgLog[msgID-1]
		anonCoordinator.DecryptedReputationMap[targetNym] = anonCoordinator.DecryptedReputationMap[targetNym] + vote
		// generate reply msg to client
		pm = map[string]interface{}{
			"reply" : true,
		}
	}

	event := &proto.Event{proto.VOTE_REPLY,pm}
	// send reply to the client
	util.Send(anonCoordinator.Socket,srcAddr,util.Encode(event))
}

func handleRoundEnd() {

}

