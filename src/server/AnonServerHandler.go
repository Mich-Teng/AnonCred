package server
import (
	"net"
	"proto"
	"encoding/gob"
	"bytes"
	"util"
	"fmt"
	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/shuffle"
	"github.com/dedis/crypto/proof"
)

var srcAddr *net.UDPAddr
var anonServer *AnonServer

func Handle(buf []byte,addr *net.UDPAddr, tmpServer *AnonServer, n int) {
	// decode the whole message
	srcAddr = addr
	anonServer = tmpServer
	event := &proto.Event{}
	err := gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(event)
	util.CheckErr(err)
	switch event.EventType {
	case proto.SERVER_REGISTER_REPLY:
		handleServerRegisterReply(event.Params);
		break
	case proto.ANNOUNCEMENT:
		handleAnnouncement(event.Params);
		break
	default:
		fmt.Println("Unrecognized request")
		break
	}
}

func handleAnnouncement(params map[string]interface{}) {
	var g = anonServer.Suite.Point()
	keyList := util.ProtobufDecodePointList(params["keys"].([]byte))
	valList := util.ProtobufDecodePointList(params["vals"].([]byte))
	if val, ok := params["g"]; ok {
		// contains g
		byteG := val.([]byte)
		g.UnmarshalBinary(byteG)
		// verify the previous shuffle
		// get the key list before shuffle
		prevKeyList := 	util.ProtobufDecodePointList(params["prev_keys"].([]byte))
		prevValList := util.ProtobufDecodePointList(params["prev_vals"].([]byte))
		verifier := shuffle.Verifier(anonServer.Suite, nil, anonServer.PublicKey, prevKeyList,
			prevValList, keyList, valList)
		err := proof.HashVerify(anonServer.Suite, "PairShuffle", verifier, params["proof"].([]byte))
		if err != nil {
			panic("Shuffle verify failed: " + err.Error())
		}
	}

	// encrypt g by modPow
	g = g.Mul(g,anonServer.PrivateKey)

	size := len(keyList)
	newKeys := make([]abstract.Point,size)
	newVals := make([]abstract.Point,size)
	for i := 0 ; i < len(keyList); i++ {
		// encrypt the public key using modPow
		newKeys[i] = anonServer.Suite.Point().Mul(keyList[i],anonServer.PrivateKey)
		// decrypt the reputation using ElGamal algorithm
		newVals[i] = util.ElGamalDecrypt(anonServer.Suite, anonServer.PrivateKey, anonServer.A, valList[i])
		// update key map
		anonServer.KeyMap[newKeys[i]] = keyList[i]
	}

	// *** perform neff shuffle here ***
	rand := anonServer.Suite.Cipher(abstract.RandomKey)
	Xbar, Ybar, prover := shuffle.Shuffle(anonServer.Suite, nil, anonServer.PublicKey,
		keyList, valList, rand)
	prf, err := proof.HashProve(anonServer.Suite, "PairShuffle", rand, prover)
	util.CheckErr(err)
	// send data to the next server
	byteXbar := util.ProtobufEncodePointList(Xbar)
	byteYbar := util.ProtobufEncodePointList(Ybar)
	byteNewKeys := util.ProtobufEncodePointList(newKeys)
	byteNewVals := util.ProtobufEncodePointList(newVals)
	byteG, _ := g.MarshalBinary()
	// prev keys means the key before shuffle
	pm := map[string]interface{}{
		"keys" : byteXbar,
		"vals" : byteYbar,
		"g" :  byteG,
		"proof" : prf,
		"prev_keys": byteNewKeys,
		"prev_vals": byteNewVals,
	}
	event := &proto.Event{proto.ANNOUNCEMENT,pm}
	util.Send(anonServer.Socket,anonServer.NextHop,util.Encode(event))
}

// handle server register reply
func handleServerRegisterReply(params map[string]interface{}) {
	reply := params["reply"].(bool)
	if val, ok := params["prev_server"]; ok {
		ServerAddr, _  := net.ResolveUDPAddr("udp",val.(string))
		anonServer.PreviousHop = ServerAddr
	}
	if reply {
		anonServer.IsConnected = true
	}
}