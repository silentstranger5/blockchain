package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
)

type UTXOSet map[string][]*TxOut

func (u *UTXOSet) Index(bc *Blockchain) {
	*u = make(map[string][]*TxOut)
	unspent := bc.UnspentTxOuts()
	for txHashStr, outs := range unspent {
		for _, out := range outs {
			(*u)[txHashStr] = append((*u)[txHashStr], out)
		}
	}
}

func (u *UTXOSet) Update(tx *Tx) {
	txHashStr := fmt.Sprintf("%x", tx.Hash())
	for _, in := range tx.TxIn {
		txOutHashStr := fmt.Sprintf("%x", in.TxOutHash)
		newOuts := make([]*TxOut, 0)
		for idx, out := range (*u)[txOutHashStr] {
			if idx != in.TxOutIndex {
				newOuts = append(newOuts, out)
			}
		}
		if len(newOuts) > 0 {
			(*u)[txOutHashStr] = newOuts
		} else {
			delete(*u, txOutHashStr)
		}
	}
	newOuts := make([]*TxOut, 0)
	for _, out := range tx.TxOut {
		newOuts = append(newOuts, out)
	}
	(*u)[txHashStr] = newOuts
}

func (u *UTXOSet) UnspentTxOuts(w *Wallet) []*TxOut {
	unspent := make([]*TxOut, 0)
	for _, outs := range *u {
		for _, out := range outs {
			if out.LockedWith(w) {
				unspent = append(unspent, out)
			}
		}
	}
	return unspent
}

func (u *UTXOSet) SpendableTxOuts(w *Wallet, amount int) (map[string][]int, int) {
	unspent := make(map[string][]int)
	total := 0
	for txHashStr, outs := range *u {
		for idx, out := range outs {
			if out.LockedWith(w) {
				unspent[txHashStr] = append(unspent[txHashStr], idx)
				total += out.Value
			}
			if total >= amount {
				return unspent, total
			}
		}
	}
	return nil, 0
}

func (u *UTXOSet) TransferTxIn(from *Wallet, amount int) ([]*TxIn, int) {
	spendable, total := u.SpendableTxOuts(from, amount)
	if len(spendable) == 0 || total < amount {
		return nil, 0
	}
	inputs := make([]*TxIn, 0)
	for txHashStr := range spendable {
		txHash, err := hex.DecodeString(txHashStr)
		if err != nil {
			panic(err)
		}
		for _, idx := range spendable[txHashStr] {
			inputs = append(inputs, &TxIn{txHash, idx, nil, nil})
		}
	}
	return inputs, total
}

func (u *UTXOSet) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(u)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func UTXOSetDeserialize(data []byte) *UTXOSet {
	u := &UTXOSet{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(u)
	if err != nil {
		panic(err)
	}
	return u
}
