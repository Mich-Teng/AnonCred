package server
import (
	"net"
	"proto"
	"encoding/gob"
	"bytes"
	"util"
	"fmt"
)

var srcAddr *net.UDPAddr
var anonServer *AnonServer

func Handle(buf []byte,addr *net.UDPAddr, tmpServer *AnonServer, n int) {
	// decode the whole message
	event := &proto.Event{}
	err := gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(event)
	util.CheckErr(err)
	switch event.EventType {
	case proto.SERVER_REGISTER_REPLY:
		handleServerRegisterReply();
		break
	default:
		fmt.Println("Unrecognized request")
		break
	}
}

func handleServerRegisterReply() {
	fmt.Println("register success")
}