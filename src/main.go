package main
import (

	"util"
	"fmt"
)
type Message struct {
	Nym     map[string][]byte
}


func main() {
	var a int = 0
	arr := util.IntToByte(a)
	aw := util.ByteToInt(arr)
	fmt.Println(aw)
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
