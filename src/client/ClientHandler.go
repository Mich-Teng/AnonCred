package client
import (
	"net"
	"encoding/gob"
	"../proto"
	"fmt"
	"../util"
	"bytes"
	"strconv"
)

func Handle(buf []byte,addr *net.UDPAddr, dissentClient *DissentClient, n int) {
	// decode the whole message
	event := &proto.Event{}
	err := gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(event)
	util.CheckErr(err)
	switch event.EventType {
	case proto.CLIENT_REGISTER_CONFIRMATION:
		handleRegisterConfirmation(dissentClient);
		break
	case proto.ANNOUNCEMENT:
		handleAnnouncement(event.Params,dissentClient);
		break
	case proto.MESSAGE:
		handleMsg(event.Params, dissentClient)
		break
	case proto.VOTE:
		handleVotePhaseStart(dissentClient)
		break
	case proto.ROUND_END:
		handleRoundEnd(dissentClient)
		break
	case proto.VOTE_REPLY:
		handleVoteReply(event.Params)
		break
	default:
		fmt.Println("Unrecognized request")
		break
	}

}


// print out register success info
func handleRegisterConfirmation(dissentClient *DissentClient) {
	dissentClient.Status = CONNECTED
	// simply print out register success info here
	fmt.Println("[client] Register success. Waiting for new round begin...");
}

// handle vote start event
func handleVotePhaseStart(dissentClient *DissentClient) {
	if dissentClient.Status != MESSAGE {
		return
	}
	// print out info in client side
	fmt.Println("*** [client] Vote Phase begins. Vote using the format... ***");
	fmt.Println("vote <msg_id> (+-)1");
}

// reset the status and prepare for the new round
func handleRoundEnd(dissentClient *DissentClient) {
	dissentClient.Status = CONNECTED
	fmt.Println("[client] Round ended. Waiting for new round start...");
}

// handle vote reply
func handleVoteReply(params map[string]interface{}) {
	status := params["reply"].(bool)
	if status == true {
		fmt.Println("Vote success!");
	}else {
		fmt.Println("Failure. Duplicate vote or verification fails!");
	}
}

// set one-time pseudonym and g, and print out info
func handleAnnouncement(params map[string]interface{}, dissentClient *DissentClient) {
	// set One-time pseudonym and g
	g := dissentClient.Suite.Point()
	// deserialize g and calculate nym

	g.UnmarshalBinary(params["g"].([]byte))
	nym := dissentClient.Suite.Point().Mul(g,dissentClient.PrivateKey)
	// set client's parameters
	dissentClient.Status = MESSAGE
	dissentClient.G = g
	dissentClient.OnetimePseudoNym = nym

	// print out the msg to suggest user to send msg or vote
	fmt.Println("[client] One-Time pseudonym for this round is ");
	fmt.Println(nym);
	fmt.Println("*** [client] Message Phase begins. Sending msg using the format... ***");
	fmt.Println("msg <msg_text>");
}

// receive the One-time pseudonym, reputation, and msg from server side
func handleMsg(params map[string]interface{}, dissentClient *DissentClient) {

	// get the reputation
	rep := params["rep"].(int)
	// get One-time pseudonym
	byteNym := params["nym"].([]byte)
	nym := dissentClient.Suite.Point()
	nym.UnmarshalBinary(byteNym)
	// get msg text
	text := params["text"].(string)
	// get msg id
	msgID := params["msgID"].(int)
	// print out in client side
	fmt.Print("Message from ")
	fmt.Print(nym)
	fmt.Println(" (reputation: " + strconv.Itoa(rep) + ")");
	fmt.Println("Message ID: " + strconv.Itoa(msgID));
	fmt.Println(text);
	fmt.Println();

}



