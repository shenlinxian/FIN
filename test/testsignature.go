package main

import (
	"dumbo_fabric/crypto/signature/bls"

	"fmt"
	"os"

	obls "github.com/bls-go-binary/bls"
)

func main() {
	//test serialize
	blsinit()
	var blssk1 obls.SecretKey
	blssk1.SetByCSPRNG()
	skbyte := blssk1.GetHexString()
	var blssk2 obls.SecretKey
	err := blssk2.SetHexString(skbyte)
	if err != nil {
		panic(err)
	}

	keypath := "/src/dumbo_fabric/config/"
	gopath := os.Getenv("GOPATH")
	//load key
	sig := bls.NewSignature()
	sig.Init(gopath+keypath, 171, 1)

	//test single signature
	msg := []byte("test bls sign         ddddddd      aaaaaaaaddddddssssswwwwwwaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaature")
	blssig := sig.Sign(msg)
	fmt.Println("length of bls signature:", len(blssig))

	if sig.Verify(1, blssig, msg) {
		fmt.Println("bls signature test pass")
	} else {
		fmt.Println("bls signature test failed")
	}

	//test aggregate signature
	num := 171
	aggsigs := make([][]byte, num)
	var mems []int32

	for i := 1; i < 171; i += 3 {
		mems = append(mems, int32(i))
	}

	for i := 0; i < len(mems); i++ {
		sig := bls.NewSignature()
		sig.Init(gopath+keypath, 171, int(mems[i]))
		blssig := sig.Sign(msg)
		aggsigs[i] = blssig
	}

	agg := sig.BatchSignature(aggsigs, msg, mems)
	fmt.Println("length of bls aggregate signature:", len(blssig))
	if sig.BatchSignatureVerify(agg, msg, mems) {
		fmt.Println("bls aggregate signature test pass")
	} else {
		fmt.Println("bls aggregate signature test failed")
	}

}

func blsinit() {
	var g_Qcoeff []uint64
	obls.Init(obls.CurveFp254BNb)
	n := obls.GetUint64NumToPrecompute()
	g_Qcoeff = make([]uint64, n)
	var Q obls.PublicKey
	obls.BlsGetGeneratorOfPublicKey(&Q)
	obls.PrecomputeG2(g_Qcoeff, obls.CastFromPublicKey(&Q))
}

