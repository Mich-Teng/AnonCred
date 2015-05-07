package main

import (
	"fmt"
//	"log"
	"net"
//	"os"
//	"bufio"
	"encoding/gob"
	"./proto"
	 "./util"
	"./client"
//	"strings"
	"strconv"
	"github.com/dedis/crypto/nist"
	"github.com/dedis/crypto/random"
//	"time"
	"github.com/dedis/protobuf"
	"bytes"
	"bufio"
	"os"
	"strings"
	"log"
	"time"
)

var dissentClient  *client.DissentClient

// register itself to controller
func register() {
	// set the parameters to register
	bytePublicKey, _ := dissentClient.PublicKey.MarshalBinary()
	params := map[string]interface{}{
		"PublicKey": bytePublicKey,
	}
	event := &proto.Event{proto.CLIENT_REGISTER_CONTROLLERSIDE,params}
	var network bytes.Buffer
	gob.NewEncoder(&network).Encode(event)
	_,err := dissentClient.Socket.Write(network.Bytes())
	util.CheckErr(err)

}

// start listener to listen port
func startClientListener() {
	fmt.Println("[debug] Client Listener started...");
	buf := make([]byte, 4096)
	for {
		n,addr,err := dissentClient.Socket.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		client.Handle(buf,addr,dissentClient,n) // a goroutine handles conn so that the loop can accept other connections
	}
}

// send message to server
func sendMsg(text string) {
	sendSigRequest(text,proto.MESSAGE)
}

// send signatured request to server
func sendSigRequest(text string, eventType int) {
	// generate signature
	rand := dissentClient.Suite.Cipher([]byte("example"))
	sig := util.ElGamalSign(dissentClient.Suite,rand,[]byte(text),dissentClient.PrivateKey,dissentClient.G)
	// serialize Point data structure
	byteNym, _ := protobuf.Encode(dissentClient.OnetimePseudoNym)
	// wrap params
	params := map[string]interface{}{
		"text": text,
		"nym":byteNym,
		"signature":sig,
	}
	event := &proto.Event{eventType,params}
	// send to coordinator
	util.Send(dissentClient.Socket,dissentClient.CoordinatorAddr,util.Encode(event))
}

// send vote to server
func sendVote(msgID, vote int) {
	// set the parameters to register
	if vote > 0 {
		vote = 1;
	}else {
		vote = -1;
	}
	v := strconv.Itoa(vote)
	m := strconv.Itoa(msgID)
	text :=  m + ";" + v
	sendSigRequest(text,proto.VOTE)
}

// initialize crypto variables
func initServer() {
	// load controller ip and port
	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1"+":"+ "10001")
	util.CheckErr(err)
	// initialize suite
	suite := nist.NewAES128SHA256QR512()
	a := suite.Secret().Pick(random.Stream)
	A := suite.Point().Mul(nil, a)
	dissentClient = &client.DissentClient{ServerAddr,nil,client.CONFIGURATION,suite,a,A,suite.Point(),nil}
}

func main() {
	initServer()
	fmt.Println("[debug] Client started...");
	// make tcp connection to controller
	conn, err := net.DialUDP("udp", nil, dissentClient.CoordinatorAddr)
	util.CheckErr(err)
	// set socket
	dissentClient.Socket = conn
	// start Listener
	go startClientListener()
	// register itself to controller
	register()

	// wait until register successful
	for ; dissentClient.Status != client.MESSAGE ; {
		time.Sleep(500 * time.Millisecond)
	}

	// read command and process
	reader := bufio.NewReader(os.Stdin)
	Loop:
	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		commands := strings.Split(command," ")
		switch commands[0] {
		case "msg":
			sendMsg(commands[1]);
			break;
		case "vote":
			msgID,_ := strconv.Atoi(commands[1])
			vote, _ := strconv.Atoi(commands[2])
			sendVote(msgID,vote)
			break;
		case "exit":
			break Loop
		}
	}
	// close connection
	conn.Close()
	fmt.Println("[debug] Exit system...");

}