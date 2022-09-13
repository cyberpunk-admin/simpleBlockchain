package blockchain

import (
	"bytes"
	"github.com/simpleBlockchain/base58"
)

type TXOutput struct {
	Value int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte)  {
	pubKeyBigInt, _:= base58.DecodeToBig(address)
	pubKeyHash := pubKeyBigInt.Bytes()
	pubKeyHash = pubKeyHash[1: len(pubKeyHash) - addressChecksumLen]

	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}