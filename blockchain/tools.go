package blockchain

import (
	"github.com/simpleBlockchain/base58"
	"log"
	"os"
	"strconv"
)

func IntToHex(x int64) []byte {
	return []byte(strconv.FormatInt(x, 16))
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func getPubKeyHashFromAddress(address string) []byte {
	pubKeyBigInt, err:= base58.DecodeToBig([]byte(address))
	if err != nil {
		log.Panic(err)
	}
	pubKeyHash := pubKeyBigInt.Bytes()
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	return pubKeyHash
}