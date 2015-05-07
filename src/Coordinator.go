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
)

var anonCoordinator *coordinator.Coordinator

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
	fmt.Println("[controller] Servers in the current network:")




}