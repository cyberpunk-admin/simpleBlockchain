package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panicf("open blockchain file err %s\n", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Printf("create blot db bucket failed: %s\n", err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Printf("put genesis block to bucket failed: %s\n", err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Printf("l command failed %s\n", err)
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

// GetBlockchain get a blockchain
func GetBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panicf("open blockchain file err %s\n", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}

// MineBlock create new block with the provided transactions
func (bc *Blockchain) MineBlock(txs []*Transaction) {
	var lastHash []byte

	for _, tx := range txs {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	_ = bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	newBlock := NewBlock(lastHash, txs)

	_ = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
}

func (bc *Blockchain) FindUTXOs(pubkeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTXs := bc.FindUnspentTransactions(pubkeyHash)
	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubkeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindUnspentTransactions(pubkeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOutputs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// check if the output spent
				if spentTXOutputs[txID] != nil {
					for _, spentOut := range spentTXOutputs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubkeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbaseTX() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubkeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOutputs[inTxID] = append(spentTXOutputs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubkeyHash)
	accumulate := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubkeyHash)&& accumulate < amount {
				accumulate += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}

			if accumulate >= amount {
				break Work
			}
		}
	}

	return accumulate, unspentOutputs
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Print(err)
			return
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Print(err)
			return false
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) Close() {
	err := bc.db.Close()
	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.db}
	return bci
}
