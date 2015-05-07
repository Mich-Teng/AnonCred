package main

import (

	"net"
	"fmt"
	"./util"
	"./coordinator"
	"bufio"
	"os"
	"github.com/dedis/crypto/nist"
	"github.com/dedis/crypto/random"
	"time"
	"log"
	"github.com/dedis/crypto/abstract"
	"proto"
)

var anonCoordinator *coordinator.Coordinator

// start server listener to read data
func startServerListener() {
	fmt.Println("[debug] Coordinator server listener started...");
	buf := make([]byte, 4096)
	for {
		n,addr,err := anonCoordinator.Socket.ReadFromUDP(buf)
		util.CheckErr(err)
		coordinator.Handle(buf,addr,anonCoordinator,n)
	}
}

// initialize coordinator
func initCoordinator() {
	ServerAddr,err := net.ResolveUDPAddr("udp",":10001")
	util.CheckErr(err)
	suite := nist.NewAES128SHA256QR512()
	a := suite.Secret().Pick(random.Stream)
	A := suite.Point().Mul(nil, a)

	anonCoordinator = &coordinator.Coordinator{ServerAddr,nil,nil,
		coordinator.CONFIGURATION,suite,a,A,nil, make(map[abstract.Point]*net.UDPAddr),
		make(map[abstract.Point]abstract.Point), nil, nil, make(map[abstract.Point]int)}
}


func clearBuffer() {
	// clear buffer
	anonCoordinator.NewClientsBuffer = nil
	// msg sender's record nym
	anonCoordinator.MsgLog = nil
}

// send the announcement notification to first server
func announce() {
	firstServer := anonCoordinator.GetFirstServer()
	if firstServer == nil {
		anonCoordinator.Status = coordinator.MESSAGE
		return
	}
	// construct reputation list (public & encrypted reputation)
	size := len(anonCoordinator.ReputationMap)
	keys := make([]abstract.Point,size)
	vals := make([]abstract.Point,size)
	i := 0
	for k, v := range anonCoordinator.ReputationMap {
		keys[i] = k
		vals[i] = v
		i++
	}
	byteKeys := util.ProtobufEncodePointList(keys)
	byteVals := util.ProtobufEncodePointList(vals)
	params := map[string]interface{}{
		"keys" : byteKeys,
		"vals" : byteVals,
	}
	event := &proto.Event{proto.ANNOUNCEMENT,params}
	util.Send(anonCoordinator.Socket,firstServer,util.Encode(event))
}

// send round end signal and data to last server
func roundEnd() {
	lastServer := anonCoordinator.GetLastServer()
	if lastServer == nil {
		anonCoordinator.Status = coordinator.READY_FOR_NEW_ROUND
		return
	}
	// add new clients into reputation map

	for _,nym := range anonCoordinator.NewClientsBuffer {
		anonCoordinator.DecryptedReputationMap[nym] = 0
	}
	// construct the parameters
	size := len(anonCoordinator.DecryptedReputationMap)
	keys := make([]abstract.Point,size)
	vals := make([]int,size)
	i := 0
	for k, v := range anonCoordinator.DecryptedReputationMap {
		keys[i] = k
		vals[i] = v
		i++
	}
	byteKeys := util.ProtobufEncodePointList(keys)
	// send signal to server
	pm := map[string]interface{} {
		"keys" : byteKeys,
		"vals" : vals,
		"no_shuffle" : false,
	}
	event := &proto.Event{proto.ROUND_END,pm}
	util.Send(anonCoordinator.Socket,lastServer,util.Encode(event))
	anonCoordinator.Status = coordinator.READY_FOR_NEW_ROUND
}

func main() {
	// init coordinator
	initCoordinator()
	// bind to socket
	conn, err := net.ListenUDP("udp",anonCoordinator.LocalAddr )
	util.CheckErr(err)
	anonCoordinator.Socket = conn
	// start listener
	go startServerListener()
	fmt.Println("** Note: Type ok to finish the server configuration. **")
	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "ok" {
			break
		}
	}
	fmt.Println("[coordinator] Servers in the current network:")
	fmt.Println(anonCoordinator.ServerList)
	anonCoordinator.Status = coordinator.READY_FOR_NEW_ROUND
	for {
		for i := 0; i < 100; i++ {
			if anonCoordinator.Status == coordinator.READY_FOR_NEW_ROUND {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
		clearBuffer()
		fmt.Println("******************** New round begin ********************")
		if anonCoordinator.Status != coordinator.READY_FOR_NEW_ROUND {
			log.Fatal("Fails to be ready for the new round")
			os.Exit(1)
		}
		anonCoordinator.Status = coordinator.ANNOUNCE
		fmt.Println("[controller] Announcement phase started...")
		announce()
		for i := 0; i < 100; i++ {
			if anonCoordinator.Status == coordinator.MESSAGE {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
		if anonCoordinator.Status != coordinator.MESSAGE {
			log.Fatal("Fails to be ready for message phase")
			os.Exit(1)
		}
		fmt.Println("[coordinator] Messaging phase started...")
		// 10 secs for msg
		time.Sleep(10000 * time.Millisecond)
		fmt.Println("[controller] Voting phase started...")
		// 10 secs for vote
		time.Sleep(10000 * time.Millisecond)
		roundEnd()
	}
}