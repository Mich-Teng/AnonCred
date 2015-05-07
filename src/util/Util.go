package util
import (
	"log"
	"bytes"
	"crypto/cipher"
	"errors"
	"github.com/dedis/crypto/abstract"
	"encoding/binary"
	"fmt"
	"os"
	"net"
	"encoding/gob"
)

func Encode(event interface{}) []byte {
	var network bytes.Buffer
	gob.NewEncoder(&network).Encode(event)
	return network.Bytes()
}

func Send(conn *net.UDPConn, addr *net.UDPAddr,content []byte) {
	_,err := conn.WriteToUDP(content, addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func SendToCoodinator(conn *net.UDPConn, content []byte) {
	_,err := conn.Write(content)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal("decode fails")
		os.Exit(1)
	}
}

func ByteToInt(b []byte) int {
	buf := bytes.NewBuffer(b) // b is []byte
	myInt, _ := binary.ReadVarint(buf)
	return int(myInt)
}

// crypto

// A basic, verifiable signature
type basicSig struct {
	C abstract.Secret // challenge
	R abstract.Secret // response
}

// Returns a secret that depends on on a message and a point
func hashElGamal(suite abstract.Suite, message []byte, p abstract.Point) abstract.Secret {
	pb, _ := p.MarshalBinary()
	c := suite.Cipher(pb)
	c.Message(nil, nil, message)
	return suite.Secret().Pick(c)
}

// This simplified implementation of ElGamal Signatures is based on
// crypto/anon/sig.go
// The ring structure is removed and
// The anonimity set is reduced to one public key = no anonimity
func ElGamalSign(suite abstract.Suite, random cipher.Stream, message []byte,
privateKey abstract.Secret, g abstract.Point) []byte {

	// Create random secret v and public point commitment T
	v := suite.Secret().Pick(random)
	T := suite.Point().Mul(g, v)

	// Create challenge c based on message and T
	c := hashElGamal(suite, message, T)

	// Compute response r = v - x*c
	r := suite.Secret()
	r.Mul(privateKey, c).Sub(v, r)

	// Return verifiable signature {c, r}
	// Verifier will be able to compute v = r + x*c
	// And check that hashElgamal for T and the message == c
	buf := bytes.Buffer{}
	sig := basicSig{c, r}
	abstract.Write(&buf, &sig, suite)
	return buf.Bytes()
}

func ElGamalVerify(suite abstract.Suite, message []byte, publicKey abstract.Point,
signatureBuffer []byte) error {

	// Decode the signature
	buf := bytes.NewBuffer(signatureBuffer)
	sig := basicSig{}
	if err := abstract.Read(buf, &sig, suite); err != nil {
		return err
	}
	r := sig.R
	c := sig.C

	// Compute base**(r + x*c) == T
	var P, T abstract.Point
	P = suite.Point()
	T = suite.Point()
	T.Add(T.Mul(nil, r), P.Mul(publicKey, c))

	// Verify that the hash based on the message and T
	// matches the challange c from the signature
	c = hashElGamal(suite, message, T)
	if !c.Equal(sig.C) {
		return errors.New("invalid signature")
	}

	return nil
}