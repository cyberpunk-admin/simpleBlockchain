package blockchain

import "bytes"

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool{
	lockingHash := HashPublicKey(in.PubKey)

	return bytes.Compare(in.Signature, lockingHash) == 0
}
