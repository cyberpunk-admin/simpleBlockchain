package blockchain

import "github.com/boltdb/bolt"

type Iterator struct {
	currentHash []byte
	db          *bolt.DB
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

