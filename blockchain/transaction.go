package blockchain

import (
	"encoding/hex"
	"log"
)

type TXInput struct {
	TxID     []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

type TXOutput struct {
	Value int
	ScriptPubKey string
}

type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

func (t Transaction) SetID() {

}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs  []TXOutput

	acc, validOutput := bc.FindSpendAbleOutput(from, amount)
	if acc < amount {
		log.Println("not enough balance")
		return nil
	}

	// list inputs
	for txid, outs := range validOutput {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			inputs = append(inputs, TXInput{txID, out, nil, nil})
		}
	}

	// list outputs
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		// change part
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}