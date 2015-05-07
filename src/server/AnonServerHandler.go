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
	"strconv"
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
	case proto.UPDATE_NEXT_HOP:
		handleUpdateNextHop(event.Params)
		break
	case proto.CLIENT_REGISTER_SERVERSIDE:
		handleClientRegisterServerSide(event.Params)
		break
	case proto.ROUND_END:
		handleRoundEnd(event.Params)
		break
	default:
		fmt.Println("Unrecognized request")
		break
	}
}

func handleRoundEnd(params map[string]interface{}) {
	var g = anonServer.Suite.Point()
	keyList := util.ProtobufDecodePointList(params["keys"].([]byte))
	var byteValList [][]byte
	if _, ok := params["start"]; ok {
		valList := params["vals"].([]int)
		// contains start, sent by coordinator
		for i := 0; i < len(valList); i++ {
			byteValList[i] = util.IntToByte(valList[i])
		}
	} else {
		// sent by server
		// verify the previous shuffle
		valList := util.ProtobufDecodePointList(params["vals"].([]byte))
		prevKeyList := 	util.ProtobufDecodePointList(params["prev_keys"].([]byte))
		prevValList := util.ProtobufDecodePointList(params["prev_vals"].([]byte))
		verifier := shuffle.Verifier(anonServer.Suite, nil, anonServer.PublicKey, prevKeyList,
			prevValList, keyList, valList)
		err := proof.HashVerify(anonServer.Suite, "PairShuffle", verifier, params["proof"].([]byte))
		if err != nil {
			panic("Shuffle verify failed: " + err.Error())
		}
		// construct byte Val list
		for i:=0; i < len(valList); i++ {
			byteValList[i] = valList[i].Data()
		}
	}


	size := len(keyList)
	newKeys := make([]abstract.Point,size)
	newVals := make([]abstract.Point,size)
	for i := 0 ; i < len(keyList); i++ {
		// decrypt the public key
		newKeys[i] = anonServer.KeyMap[keyList[i]]
		// encrypt the reputation using ElGamal algorithm
		newVals[i] = util.ElGamalEncrypt(anonServer.Suite,anonServer.PublicKey,byteValList[i])
	}

	// *** perform neff shuffle here ***
	rand := anonServer.Suite.Cipher(abstract.RandomKey)
	Xbar, Ybar, prover := shuffle.Shuffle(anonServer.Suite, nil, anonServer.PublicKey,
		newKeys, newVals, rand)
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
		"proof" : prf,
		"prev_keys": byteNewKeys,
		"prev_vals": byteNewVals,
	}
	event := &proto.Event{proto.ANNOUNCEMENT,pm}
	util.Send(anonServer.Socket,anonServer.PreviousHop,util.Encode(event))
}

// encrypt the public key and send to next hop
func handleClientRegisterServerSide(params map[string]interface{}) {
	publicKey := anonServer.Suite.Point()
	err := publicKey.UnmarshalBinary(params["public_key"].([]byte))
	util.CheckErr(err)

	newKey := anonServer.Suite.Point().Mul(publicKey,anonServer.Roundkey)

	pm := map[string]interface{}{
		"public_key" : newKey.MarshalBinary(),
	}
	event := &proto.Event{proto.CLIENT_REGISTER_SERVERSIDE,pm}
	util.Send(anonServer.Socket,anonServer.NextHop,util.Encode(event))
	// add into key map
	anonServer.KeyMap[newKey] = publicKey
}

func handleUpdateNextHop(params map[string]interface{}) {
	addr, err := net.ResolveUDPAddr("udp",params["next_hop"].(string))
	util.CheckErr(err)
	anonServer.NextHop = addr
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
	g = g.Mul(g,anonServer.Roundkey)

	size := len(keyList)
	newKeys := make([]abstract.Point,size)
	newVals := make([]abstract.Point,size)
	for i := 0 ; i < len(keyList); i++ {
		// encrypt the public key using modPow
		newKeys[i] = anonServer.Suite.Point().Mul(keyList[i],anonServer.Roundkey)
		// decrypt the reputation using ElGamal algorithm
		newVals[i] = util.ElGamalDecrypt(anonServer.Suite, anonServer.PrivateKey, anonServer.A, valList[i])
		// update key map
		anonServer.KeyMap[newKeys[i]] = keyList[i]
	}

	// *** perform neff shuffle here ***
	rand := anonServer.Suite.Cipher(abstract.RandomKey)
	Xbar, Ybar, prover := shuffle.Shuffle(anonServer.Suite, nil, anonServer.PublicKey,
		newKeys, newVals, rand)
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