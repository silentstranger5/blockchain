package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
)

// reward is amount of mining reward assets
const reward = 10

// Transaction is a transaction structure
type Transaction struct {
	ID   string
	Vin  []TXInput
	Vout []TXOutput
}

// TXOutput is a transaction output structure
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// TXInput is a transaction input structure
type TXInput struct {
	TXID      string
	Vout      int
	Signature []byte
	PubKey    []byte
}

// TXOutput is a structure containing a collection
// of transaction outputs
type TXOutputs struct {
	Outputs []TXOutput
}

// UsesKey checks if transaction input uses the same public key
// that had been used to create the specified key hash
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

// Lock assigns public key hash to the transaction output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// LockedWith checks if the transaction output
// has specified public key hash
func (out *TXOutput) LockedWith(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// NewCoinbaseTX returns a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	txin := TXInput{"", -1, nil, []byte(data)}
	txout := *NewTXOutput(reward, to)
	tx := Transaction{"", []TXInput{txin}, []TXOutput{txout}}
	tx.ID = tx.Hash()
	return &tx
}

// NewTXOutput returns a new transaction output
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

// Hash provides a string hash of the transaction
func (tx *Transaction) Hash() string {
	data := fmt.Sprintf("%v", tx)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// NewTransaction returns a new transaction
func (bc *Blockchain) NewTransaction(from, to string, amount int, u *UTXOSet) *Transaction {
	var txinputs []TXInput
	var txoutputs []TXOutput

	wallets, err := GetWallets()
	if err != nil {
		log.Fatal(err)
	}
	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, outputs := u.FindSpendable(pubKeyHash, amount)
	if acc < amount {
		log.Fatal("Not enough funds")
	}
	for txid, outs := range outputs {
		for _, out := range outs {
			input := TXInput{txid, out, nil, wallet.PublicKey}
			txinputs = append(txinputs, input)
		}
	}
	txoutputs = append(txoutputs, *NewTXOutput(amount, to))
	if acc > amount {
		txoutputs = append(txoutputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{"", txinputs, txoutputs}
	tx.ID = tx.Hash()
	u.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// FindUnspent returns a map from transaction id to
// a slice of transaction outputs
func (bc *Blockchain) FindUnspent() map[string]TXOutputs {
	var unspent = make(map[string]TXOutputs)
	var spent = make(map[string][]int)
	var bci = bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
		Outputs:
			for outidx, out := range tx.Vout {
				if spent[tx.ID] != nil {
					for _, spentout := range spent[tx.ID] {
						if spentout == outidx {
							continue Outputs
						}
					}
				}
				outs := unspent[tx.ID]
				outs.Outputs = append(outs.Outputs, out)
				unspent[tx.ID] = outs
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					spent[in.TXID] = append(spent[in.TXID], in.Vout)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspent
}

// IsCoinbase checks if specified transaction
// is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TXID) == 0 && tx.Vin[0].Vout == -1
}

// Sign signs transaction with the provided
// ECDSA private key
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[vin.TXID]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(txCopy.ID))
		if err != nil {
			log.Fatal(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

// TrimmedCopy returns Transaction with transaction inputs
// without signatures and public keys
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.TXID, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify checks correctness of a signature of the transaction
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[vin.TXID]
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
		if !ecdsa.Verify(&rawPubKey, []byte(txCopy.ID), &r, &s) {
			return false
		}
	}
	return true
}

// Serialize returns a byte slice representation
// of the specified transaction outputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Fatal(err)
	}
	return buff.Bytes()
}

// DeserializeOutputs returns transaction outputs
// obtained from a byte slice serialization
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Fatal(err)
	}
	return outputs
}
