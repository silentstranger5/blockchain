package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type UTXOSet map[string][]*TxOut

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

func GetUTXOSet() (*UTXOSet, error) {
	_, err := os.Stat("data")
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("data", 0750)
		if err != nil {
			return nil, err
		}
	}
	_, err = os.Stat("data/utxoset.json")
	if errors.Is(err, os.ErrNotExist) {
		u := new(UTXOSet)
		*u = make(UTXOSet)
		data, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile("data/utxoset.json", data, 0666)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	data, err := os.ReadFile("data/utxoset.json")
	if err != nil {
		return nil, err
	}
	u := new(UTXOSet)
	*u = make(UTXOSet)
	err = json.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UTXOSet) Write() error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/utxoset.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}
