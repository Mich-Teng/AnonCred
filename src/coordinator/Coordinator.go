package coordinator
import (
	"github.com/dedis/crypto/abstract"
	"net"
)

type Coordinator struct {
	// local address
	LocalAddr *net.UDPAddr
	// socket
	Socket *net.UDPConn
	// network topology for server cluster
	ServerList []*net.UDPAddr
	// initialize the controller status
	Status int


	// crypto things
	Suite abstract.Suite
	// private key
	PrivateKey abstract.Secret
	// public key
	PublicKey abstract.Point
	// generator g
	G abstract.Point

	Clients map[abstract.Point]*net.UDPAddr
	/*
	// message sender list
	MsgSenderList list.List
	// collect vote for this round
	VoteCollect map[abstract.Point]int
	// buffer for new clients joined during this round
	NewClientBuffer list.List
	// log info for all the votes
	VoteLog list.List
	*/

}

// get last server in topology
func (c Coordinator) GetLastServer() *net.UDPAddr {
	return c.ServerList[len(c.ServerList)-1]
}

// get first server in topology
func (c Coordinator) GetFirstServer() *net.UDPAddr {
	return c.ServerList[0]
}

func (c Coordinator) AddClient(key abstract.Point, val *net.UDPAddr) {
	c.Clients[key] = val
}