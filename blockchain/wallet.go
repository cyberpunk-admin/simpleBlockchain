package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"math/big"

	"github.com/simpleBlockchain/base58"

	"golang.org/x/crypto/ripemd160"
)

const (
	version = byte(0x00)
	addressChecksumLen = 4
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

func (w *Wallet) GetAddress() []byte {
	var address []byte

	publicKeyHash := HashPublicKey(w.PublicKey)

	versionedPayload := append(publicKeyHash, []byte{version}...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)

	address = base58.EncodeBig(address, big.NewInt(0).SetBytes(fullPayload))
	return address
}

// ValidateAddress check if address is valid
func ValidateAddress(address string) bool {
	pubKeyBigInt, _:= base58.DecodeToBig([]byte(address))
	pubKeyHash := pubKeyBigInt.Bytes()

	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]

	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func HashPublicKey(pubKey []byte) []byte {
	pubKeySHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(pubKeySHA256[:])
	if err != nil {
		log.Panic(err)
	}

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}


func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()

	return &Wallet{privateKey, publicKey}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	publicKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, publicKey
}

func checksum(payload []byte) []byte {
	// two sha256 check sum
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}