package main
import (

	"fmt"
	"github.com/dedis/crypto/nist"
	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/random"
	"util"
)
type Message struct {
	Nym     map[string][]byte
}


func main() {
	suite1 := nist.NewAES128SHA256QR512()
	l := make([]abstract.Point,1)
	key := suite1.Secret().Pick(random.Stream)
	l[0] = suite1.Point().Mul(nil,key)
	m := make(map[string]int)
	m[l[0].String()] = 2
	bytes := util.ProtobufEncodePointList(l)
	keyList := util.ProtobufDecodePointList(bytes)

	fmt.Println(m[l[0].String()])
	fmt.Println(m[keyList[0].String()])
	fmt.Println(l[0].String())
	fmt.Println(keyList[0].String())
	/*
	var a int = -1
	c := util.IntToByte(a)
	b := util.ByteToInt(c)
	fmt.Println(b)
	*/
	/*
	suite1 := nist.NewAES128SHA256QR512()
	suite2 := nist.NewAES128SHA256QR512()
	suite3 := nist.NewAES128SHA256QR512()
	key := suite1.Secret().Pick(random.Stream)
	g1 := suite1.Point().Mul(nil,key)
	byte1, _ := g1.MarshalBinary()
	fmt.Println(g1)

	g2 := suite2.Point()
	err := g2.UnmarshalBinary(byte1)
	g2 = suite2.Point().Mul(g2,key)
	bytes2, _ := g2.MarshalBinary()
	fmt.Println(g2)

	g3 := suite3.Point()
	err = g3.UnmarshalBinary(bytes2)
	util.CheckErr(err)
	g3 = suite3.Point().Mul(g3,key)

	fmt.Println(g3)
	*/

	/*
	var aSecret abstract.Secret
	var tSecret = reflect.TypeOf(&aSecret).Elem()

	suite := nist.NewAES128SHA256QR512()
	cons := protobuf.Constructors {
		tSecret: func()interface{} { return suite.Secret() },
	}

	a := suite.Secret().Pick(random.Stream)
	b := suite.Secret().Pick(random.Stream)
	fmt.Println(a)
	fmt.Println(b)

	byteA, _ := a.MarshalBinary()
	byteB, _ := b.MarshalBinary()
	l := map[string][]byte {
		"a":byteA,
		"b":byteB,
	}

	byteNym, err := protobuf.Encode(&Message{l})
	if err != nil {
		fmt.Println(err.Error())
	}

	var msg Message
	if err = protobuf.DecodeWithConstructors(byteNym, &msg, cons); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(msg.Nym["a"])
	fmt.Println(msg.Nym["b"])
	*/
}
