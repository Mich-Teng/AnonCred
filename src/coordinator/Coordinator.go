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
	if len(c.ServerList) == 0 {
		return nil
	}
	return c.ServerList[len(c.ServerList)-1]
}

// get first server in topology
func (c Coordinator) GetFirstServer() *net.UDPAddr {
	if len(c.ServerList) == 0 {
		return nil
	}
	return c.ServerList[0]
}

func (c Coordinator) AddClient(key abstract.Point, val *net.UDPAddr) {
	// delete the client who has same ip address
	for k,v := range c.Clients {
		if v.String() == val.String() {
			delete(c.Clients,k)
			break
		}
	}
	c.Clients[key] = val
}

// get first server in topology
func (c Coordinator) AddServer(addr *net.UDPAddr){
	c.ServerList = append(c.ServerList,addr)
}