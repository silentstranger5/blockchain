package blockchain

import (
	"bytes"
)

type Tx struct {
	From   *Wallet
	To     *Wallet
	Amount int64
}

type Txs []*Tx

func (txs Txs) Bytes() ([]byte, error) {
	var result = make([]byte, 0)
	for _, tx := range txs {
		txbytes, err := tx.Bytes()
		if err != nil {
			return nil, err
		}
		result = append(result, txbytes...)
	}
	return result, nil
}

func (tx *Tx) Bytes() ([]byte, error) {
	txambytes, err := IntToBytes(tx.Amount)
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{
		tx.From.Bytes(),
		tx.To.Bytes(),
		txambytes,
	}, []byte{}), nil
}
