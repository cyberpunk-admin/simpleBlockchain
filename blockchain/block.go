package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce		  int64
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

func DeserializeBlock(b []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&block)
	if err != nil {
		log.Printf("decoder err: %s\n", err)
	}
	return &block
}

func NewBlock(prevBlockHash []byte, Data string) *Block {
	block := &Block{Timestamp: time.Now().Unix(), Data: []byte(Data), PrevBlockHash: prevBlockHash}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}
