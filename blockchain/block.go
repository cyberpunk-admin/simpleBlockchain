package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)


type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int64
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	if err != nil {
		log.Printf("encoder err: %s\n", err)
	}
	return res.Bytes()
}

func (b *Block) HashTransaction() []byte {
	var txHashs [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashs = append(txHashs, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashs, []byte{}))
	return txHash[:]
}

func DeserializeBlock(b []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&block)
	if err != nil {
		log.Printf("decoder err: %s\n", err)
	}
	return &block
}

func NewBlock(prevBlockHash []byte, Txs []*Transaction) *Block {
	block := &Block{Timestamp: time.Now().Unix(), Transactions: Txs, PrevBlockHash: prevBlockHash}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]byte{}, []*Transaction{coinbase})
}

