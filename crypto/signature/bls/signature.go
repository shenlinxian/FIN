package bls

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/bls-go-binary/bls"
)

type Signature struct {
	sk     []byte
	pks    [][]byte
	blssk  bls.SecretKey
	blspks []bls.PublicKey
}

func (sig *Signature) Sign(msg []byte) []byte {
	h := sha256.New()
	h.Write(msg)
	hashText := h.Sum(nil)

	blssig := sig.blssk.Sign(string(hashText))
	return []byte(blssig.GetHexString())
}

func (sig *Signature) Verify(id int, signature []byte, msg []byte) bool {
	h := sha256.New()
	h.Write(msg)
	hashText := h.Sum(nil)

	var blssig bls.Sign
	err := blssig.SetHexString(string(signature))
	if err != nil {
		panic(err)
	}
	return blssig.Verify(&sig.blspks[id], string(hashText))
}

//unify half-aggrate schnorr signatures and simply join ecdsa and schnorr signature
func (sig *Signature) BatchSignature(signatures [][]byte, msg []byte, mems []int32) []byte {
	var agg bls.Sign
	sigs := make([]bls.Sign, len(mems))
	for i := 0; i < len(mems); i++ {
		var blssig bls.Sign
		err := blssig.SetHexString(string(signatures[i]))
		if err != nil {
			panic(err)
		}
		sigs[i] = blssig
	}
	agg.Aggregate(sigs)
	return []byte(agg.GetHexString())
}

//unify half-aggrate schnorr signatures verify and simply verify ecdsa or schnorr signature one by one
func (sig *Signature) BatchSignatureVerify(BatchSignature []byte, msg []byte, mems []int32) bool {
	h := sha256.New()
	h.Write(msg)
	hashText := h.Sum(nil)

	blspks := make([]bls.PublicKey, len(mems))
	for i := 0; i < len(mems); i++ {

		blspks[i] = sig.blspks[mems[i]-1]
	}

	var aggsig bls.Sign
	err := aggsig.SetHexString(string(BatchSignature))
	if err != nil {
		panic(err)
	}

	return aggsig.FastAggregateVerify(blspks, hashText)
}

//load keys
func (sig *Signature) Init(path string, num int, id int) {
	//init bls
	var g_Qcoeff []uint64
	bls.Init(bls.CurveFp254BNb)
	n := bls.GetUint64NumToPrecompute()
	g_Qcoeff = make([]uint64, n)
	var Q bls.PublicKey
	bls.BlsGetGeneratorOfPublicKey(&Q)
	bls.PrecomputeG2(g_Qcoeff, bls.CastFromPublicKey(&Q))
	//load sk
	skpath := fmt.Sprintf("%s/blskey/sk/sk%d.pem", path, id-1)
	file, err := os.Open(skpath)
	if err != nil {
		panic(err)
	}
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, info.Size())
	file.Read(buf)
	block, _ := pem.Decode(buf)
	sig.sk = block.Bytes
	file.Close()

	var sk bls.SecretKey
	err = sk.SetHexString(string(sig.sk))
	if err != nil {
		panic(err)
	}
	sig.blssk = sk

	//load pk
	sig.pks = make([][]byte, num)
	sig.blspks = make([]bls.PublicKey, num)
	for i := 0; i < num; i++ {
		pkpath := fmt.Sprintf("%s/blskey/pk/pk%d.pem", path, i)
		file, err := os.Open(pkpath)
		if err != nil {
			panic(err)
		}
		info, err := file.Stat()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, info.Size())
		file.Read(buf)
		block, _ := pem.Decode(buf)
		sig.pks[i] = block.Bytes
		file.Close()

		var blspk bls.PublicKey
		err = blspk.SetHexString(string(sig.pks[i]))
		if err != nil {
			panic(err)
		}
		sig.blspks[i] = blspk
	}

}

func NewSignature() Signature {
	return Signature{}
}

