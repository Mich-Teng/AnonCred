package main
import (
	"fmt"
	"net"
	"util"
	"./server"
	"time"
	"log"
	"os"
	"github.com/dedis/crypto/nist"
	"github.com/dedis/crypto/random"
	"github.com/dedis/crypto/abstract"
	"./proto"
)

var anonServer *server.AnonServer

// register itself to controller
func serverRegister() {
	// set the parameters to register
	params := map[string]interface{}{}
	event := &proto.Event{proto.SERVER_REGISTER,params}

	util.SendToCoodinator(anonServer.Socket,util.Encode(event))
}

func startAnonServerListener() {
	fmt.Println("[debug] AnonServer Listener started...");
	buf := make([]byte, 4096)
	for {
		n,addr,err := anonServer.Socket.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		server.Handle(buf,addr,anonServer,n)
	}
}

func initAnonServer() {
	// load controller ip and port
	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1"+":"+ "10001")
	util.CheckErr(err)
	// initialize suite
	suite := nist.NewAES128SHA256QR512()
	a := suite.Secret().Pick(random.Stream)
	A := suite.Point().Mul(nil, a)
	RoundKey, err := suite.Secret().Pick(random.Stream)
	util.CheckErr(err)
	anonServer = &server.AnonServer{ServerAddr,nil,suite,a,A,suite.Point(),nil,
	false,nil,nil,make(map[abstract.Point]abstract.Point),nil,RoundKey}
}

func main() {
	// init anon server
	initAnonServer()
	fmt.Println("[debug] AnonServer started...");
	// make tcp connection to controller
	conn, err := net.DialUDP("udp", nil, anonServer.CoordinatorAddr)
	util.CheckErr(err)
	// set socket
	anonServer.Socket = conn
	// start Listener
	go startAnonServerListener()
	// register itself to coordinator
	serverRegister()

	// wait until register successful
	for i := 0 ; i < 100 ; i++ {
		if anonServer.IsConnected {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if anonServer.IsConnected != true {
		log.Fatal("Fails to connect to coordinator")
		os.Exit(1)
	}
	fmt.Println("[debug] Register success...")
	for {
		time.Sleep(100000000 * time.Millisecond)
	}

	fmt.Println("[debug] Exit system...");

}
