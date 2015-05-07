package server
import (
	"net"
	"github.com/dedis/crypto/abstract"
)


type AnonServer struct {
	// client-side config
	CoordinatorAddr *net.UDPAddr
	Socket *net.UDPConn
	// crypto variables
	Suite abstract.Suite
	PrivateKey abstract.Secret
	PublicKey abstract.Point
	OnetimePseudoNym abstract.Point
	G abstract.Point

	// buffer data
	IsConnected bool
	// next hop in topology
	NextHop *net.UDPAddr
	// previous hop in topology
	PreviousHop *net.UDPAddr
	// map current public key with previous key
	KeyMap map[abstract.Point]abstract.Point
	// generated by elgmal encryption
	A abstract.Point

}