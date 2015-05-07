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

	anonCoordinator = &coordinator.Coordinator{ServerAddr,nil,make([]*net.UDPAddr,2),coordinator.CONFIGURATION,suite,a,A,nil}
}

// todo
func clearBuffer() {
	// todo
}

// todo
func announce() {
	// todo
}

// todo
func roundEnd() {
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