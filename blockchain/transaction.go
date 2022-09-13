package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

const (
	subsidy = 8
)

// Transaction represent a sampleBlockchain transaction
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Sign(PriKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTX() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, _ := ecdsa.Sign(rand.Reader, &PriKey, txCopy.ID)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// Trimmed get new input list
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	// Trimmed get a coped output list
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx *Transaction) IsCoinbaseTX() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(tx)
	if err != nil {
		log.Printf("serialize transaction error %s", err)
		return nil
	}
	return res.Bytes()
}

func DeserializeTransaction(b []byte) *Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Printf("Deserialize transaction error %s", err)
		return nil
	}
	return &transaction
}

func (tx *Transaction) SetID() {
	tx.ID = tx.Hash()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func NewUTXOTransaction(wallet *Wallet, to string, amount int, bc *Blockchain) *Transaction{
	var txInputs []TXInput
	var txOutputs []TXOutput

	pubKeyHash := HashPublicKey(wallet.PublicKey)

	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("No enough Founds")
	}

	for txid, validOutputIdx := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range validOutputIdx {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			txInputs = append(txInputs, input)
		}
	}

	from := fmt.Sprintf("%s", wallet.GetAddress())
	txOutputs = append(txOutputs, *NewTXOutput(amount, from))

	// Part as a change
	if acc > amount {
		txOutputs = append(txOutputs, *NewTXOutput(acc - amount, from))
	}

	tx := Transaction{nil, txInputs, txOutputs}
	tx.SetID()

	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}


// NewCoinbaseTX new a coinbase transaction
// It represents a transaction without previously existing outputs
func NewCoinbaseTX(to, data string) *Transaction{
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}

	txin := TXInput{[]byte{}, -1, nil,[]byte(data)}
	txout := *NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}