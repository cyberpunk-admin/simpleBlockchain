package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
)

const (
	dbFile = "../blockchain.db"
	blocksBucket = "blocks"
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type Iterator struct {
	currentHash []byte
	db *bolt.DB
}

func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Printf("open blockchain file err %s\n", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, _ := tx.CreateBucket([]byte(blocksBucket))
			_ = b.Put(genesis.Hash, genesis.Serialize())
			_ = b.Put([]byte("l"), genesis.Hash)
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	bc := Blockchain{tip, db}
	return &bc
}

func (bc *Blockchain) AddBlock(Data string) error{
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	newBlock := NewBlock(lastHash, Data)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash
		return err
	})
	return err
}

func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.db}
	return bci
}

func (bc *Blockchain) FindSpendAbleOutput(address string, amount int) (int, map[string][]int) {
	unspendOutputs := make(map[string][]int)
	unspendTxs := bc.FindUnspendTranscations()
	accumulated := 0

}

func (bc *Blockchain) FindUnspendTranscations() interface{} {

}

func (it *Iterator) Next() *Block {
	var block *Block
	_ = it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(it.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	it.currentHash = block.PrevBlockHash
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock([]byte{}, "first simple blockchain")
}
