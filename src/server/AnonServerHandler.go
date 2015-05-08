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
	"github.com/dedis/crypto/random"
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
	keyList := util.ProtobufDecodePointList(params["keys"].([]byte))
	size := len(keyList)
	fmt.Println(size)
	var byteValList = make([][]byte, size)
	if _, ok := params["is_start"]; ok {
		valList := params["vals"].([]int)
		// contains start, sent by coordinator
		for i := 0; i < len(valList); i++ {
			byteValList[i] = util.IntToByte(valList[i])
		}
	} else {
		// sent by server
		// verify the previous shuffle
		valList := util.ProtobufDecodePointList(params["vals"].([]byte))
		if _, shuffled := params["shuffled"]; shuffled {
			prevKeyList := util.ProtobufDecodePointList(params["prev_keys"].([]byte))
			prevValList := util.ProtobufDecodePointList(params["prev_vals"].([]byte))
			verifier := shuffle.Verifier(anonServer.Suite, nil, anonServer.PublicKey, prevKeyList,
				prevValList, keyList, valList)
			err := proof.HashVerify(anonServer.Suite, "PairShuffle", verifier, params["proof"].([]byte))
			if err != nil {
				panic("Shuffle verify failed: " + err.Error())
			}
		}
		// construct byte Val list
		for i:=0; i < len(valList); i++ {
			byteValList[i], _ = valList[i].Data()
		}
	}


	newKeys := make([]abstract.Point,size)
	newVals := make([]abstract.Point,size)
	for i := 0 ; i < len(keyList); i++ {
		// decrypt the public key
		newKeys[i] = anonServer.KeyMap[keyList[i].String()]
		fmt.Println("keymap: ")
		fmt.Println(anonServer.KeyMap)
		fmt.Println("key: ")
		fmt.Println(keyList[i])
		fmt.Println(newKeys[i])
		// encrypt the reputation using ElGamal algorithm
		K,C,_ := util.ElGamalEncrypt(anonServer.Suite,anonServer.PublicKey,byteValList[i])
		fmt.Println("[handle round end]elgamal decrypt data : ")
		fmt.Println(len(byteValList[i]))
		fmt.Println(byteValList[i])
		newVals[i] = C
		anonServer.A = K
	}
	byteNewKeys := util.ProtobufEncodePointList(newKeys)
	byteNewVals := util.ProtobufEncodePointList(newVals)

	if(size <= 1) {
		// no need to shuffle, just send the package to next server
		pm := map[string]interface{}{
			"keys" : byteNewKeys,
			"vals" : byteNewVals,
		}
		event := &proto.Event{proto.ROUND_END,pm}
		fmt.Println(anonServer.PreviousHop)
		util.Send(anonServer.Socket,anonServer.PreviousHop,util.Encode(event))
		// reset RoundKey and key map
		anonServer.Roundkey = anonServer.Suite.Secret().Pick(random.Stream)
		anonServer.KeyMap = make(map[string]abstract.Point)
		return
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

	// prev keys means the key before shuffle
	pm := map[string]interface{}{
		"keys" : byteXbar,
		"vals" : byteYbar,
		"proof" : prf,
		"prev_keys": byteNewKeys,
		"prev_vals": byteNewVals,
		"shuffled":true,
	}
	event := &proto.Event{proto.ROUND_END,pm}
	util.Send(anonServer.Socket,anonServer.PreviousHop,util.Encode(event))

	// reset RoundKey and key map
	anonServer.Roundkey = anonServer.Suite.Secret().Pick(random.Stream)
	anonServer.KeyMap = make(map[string]abstract.Point)
}

// encrypt the public key and send to next hop
func handleClientRegisterServerSide(params map[string]interface{}) {
	publicKey := anonServer.Suite.Point()
	err := publicKey.UnmarshalBinary(params["public_key"].([]byte))
	util.CheckErr(err)

	newKey := anonServer.Suite.Point().Mul(publicKey,anonServer.Roundkey)
	byteNewKey, err := newKey.MarshalBinary()
	util.CheckErr(err)
	pm := map[string]interface{}{
		"public_key" : byteNewKey,
		"addr" : params["addr"].(string),
	}
	event := &proto.Event{proto.CLIENT_REGISTER_SERVERSIDE,pm}
	util.Send(anonServer.Socket,anonServer.NextHop,util.Encode(event))
	// add into key map
	fmt.Println("[debug] Receive client register request... ")
	anonServer.KeyMap[newKey.String()] = publicKey
}

func handleUpdateNextHop(params map[string]interface{}) {
	addr, err := net.ResolveUDPAddr("udp",params["next_hop"].(string))
	util.CheckErr(err)
	anonServer.NextHop = addr
}

func handleAnnouncement(params map[string]interface{}) {
	var g abstract.Point = nil
	keyList := util.ProtobufDecodePointList(params["keys"].([]byte))
	valList := util.ProtobufDecodePointList(params["vals"].([]byte))
	size := len(keyList)

	if val, ok := params["g"]; ok {
		// contains g
		byteG := val.([]byte)
		g = anonServer.Suite.Point()
		g.UnmarshalBinary(byteG)
		g = anonServer.Suite.Point().Mul(g,anonServer.Roundkey)
		// verify the previous shuffle
		// get the key list before shuffle
		if _, shuffled := params["shuffle"]; shuffled {
			prevKeyList := 	util.ProtobufDecodePointList(params["prev_keys"].([]byte))
			prevValList := util.ProtobufDecodePointList(params["prev_vals"].([]byte))
			verifier := shuffle.Verifier(anonServer.Suite, nil, anonServer.PublicKey, prevKeyList,
				prevValList, keyList, valList)
			err := proof.HashVerify(anonServer.Suite, "PairShuffle", verifier, params["proof"].([]byte))
			if err != nil {
				panic("Shuffle verify failed: " + err.Error())
			}
		}
	}else {
		g = anonServer.Suite.Point().Mul(nil,anonServer.Roundkey)
	}



	newKeys := make([]abstract.Point,size)
	newVals := make([]abstract.Point,size)
	for i := 0 ; i < len(keyList); i++ {
		// encrypt the public key using modPow
		newKeys[i] = anonServer.Suite.Point().Mul(keyList[i],anonServer.Roundkey)
		// decrypt the reputation using ElGamal algorithm
		newVals[i] = util.ElGamalDecrypt(anonServer.Suite, anonServer.PrivateKey, anonServer.A, valList[i])
		fmt.Println("[handle announcement]elgamal decrypt data : ")
		fmt.Println(valList[i].Data())
		// update key map
		anonServer.KeyMap[newKeys[i].String()] = keyList[i]
	}
	byteNewKeys := util.ProtobufEncodePointList(newKeys)
	byteNewVals := util.ProtobufEncodePointList(newVals)
	byteG, err := g.MarshalBinary()
	util.CheckErr(err)

	if(size <= 1) {
		// no need to shuffle, just send the package to next server
		pm := map[string]interface{}{
			"keys" : byteNewKeys,
			"vals" : byteNewVals,
			"g" : byteG,
		}
		event := &proto.Event{proto.ANNOUNCEMENT,pm}
		util.Send(anonServer.Socket,anonServer.NextHop,util.Encode(event))
		return
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

	// prev keys means the key before shuffle
	pm := map[string]interface{}{
		"keys" : byteXbar,
		"vals" : byteYbar,
		"g" :  byteG,
		"proof" : prf,
		"prev_keys": byteNewKeys,
		"prev_vals": byteNewVals,
		"shuffle": true,
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