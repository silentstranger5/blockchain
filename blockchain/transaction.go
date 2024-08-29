package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"reflect"
)

type Tx struct {
	TxIn  []*TxIn
	TxOut []*TxOut
}

func (tx *Tx) Bytes() []byte {
	data := make([][]byte, 0)
	for _, txin := range tx.TxIn {
		data = append(data, txin.Bytes())
	}
	for _, txout := range tx.TxOut {
		data = append(data, txout.Bytes())
	}
	return bytes.Join(data, nil)
}

func (tx *Tx) Hash() []byte {
	hash := sha256.Sum256(tx.Bytes())
	return hash[:]
}

func (tx *Tx) Trim() *Tx {
	txcopy := new(Tx)
	*txcopy = *tx
	txcopy.TxIn = nil
	for _, txIn := range tx.TxIn {
		v := *txIn
		txcopy.TxIn = append(txcopy.TxIn, &v)
	}
	for _, in := range txcopy.TxIn {
		in.Signature = nil
		in.PubKey = nil
	}
	return txcopy
}

func (tx *Tx) Sign(w *Wallet) {
	privateKey := (*ecdsa.PrivateKey)(w)
	publicKey := (*ecdsa.PublicKey)(&w.PublicKey)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, tx.Trim().Hash())
	if err != nil {
		panic(err)
	}
	for _, in := range tx.TxIn {
		in.Signature = signature
		in.PubKey = append(
			publicKey.X.Bytes(),
			publicKey.Y.Bytes()...,
		)
	}
}

func (tx *Tx) SignedWith(w *Wallet) bool {
	walletPubKey := (*ecdsa.PublicKey)(&w.PublicKey)
	ok := true
	for _, in := range tx.TxIn {
		ok = ok && ecdsa.VerifyASN1(walletPubKey, tx.Trim().Hash(), in.Signature)
	}
	return ok
}

func (tx *Tx) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(tx)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TxDeserialize(data []byte) *Tx {
	tx := &Tx{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(tx)
	if err != nil {
		panic(err)
	}
	return tx
}

type TxIn struct {
	TxOutHash  []byte
	TxOutIndex int
	Signature  []byte
	PubKey     []byte
}

func (in *TxIn) Bytes() []byte {
	return bytes.Join([][]byte{
		in.TxOutHash,
		IntToBytes(in.TxOutIndex),
		in.Signature,
		in.PubKey,
	}, nil)
}

type TxOut struct {
	Value      int
	PubKeyHash []byte
}

func (out *TxOut) Bytes() []byte {
	return append(IntToBytes(out.Value), out.PubKeyHash...)
}

func (out *TxOut) LockedWith(w *Wallet) bool {
	return reflect.DeepEqual([]byte(out.PubKeyHash), w.PubKeyHash())
}

type Txs []*Tx

func (txs Txs) Bytes() []byte {
	data := make([][]byte, 0)
	for _, tx := range txs {
		data = append(data, tx.Bytes())
	}
	return bytes.Join(data, nil)
}

func (txs Txs) Hash() []byte {
	hash := sha256.Sum256(txs.Bytes())
	return hash[:]
}

func (txs Txs) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(txs)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TxsDeserialize(data []byte) *Txs {
	txs := &Txs{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(txs)
	if err != nil {
		panic(err)
	}
	return txs
}

func (txs *Txs) UnspentTxOuts(spent map[string][]int) map[string][]*TxOut {
	unspent := make(map[string][]*TxOut)
	for _, tx := range *txs {
		txHashStr := fmt.Sprintf("%x", tx.Hash())
		for _, in := range tx.TxIn {
			txOutHashStr := fmt.Sprintf("%x", in.TxOutHash)
			spent[txOutHashStr] = append(spent[txOutHashStr], in.TxOutIndex)
		}
		for outIdx, out := range tx.TxOut {
			spentout := false
			for _, spentIdx := range spent[txHashStr] {
				if spentIdx == outIdx {
					spentout = true
					break
				}
			}
			if spentout {
				continue
			}
			unspent[txHashStr] = append(unspent[txHashStr], out)
		}
	}
	return unspent
}

func CoinBaseTx(wallet *Wallet) *Tx {
	txin := []*TxIn{&TxIn{}}
	txout := []*TxOut{&TxOut{reward, wallet.PubKeyHash()}}
	tx := &Tx{txin, txout}
	tx.Sign(wallet)
	return tx
}

func TransferTx(from, to *Wallet, amount int, u *UTXOSet) *Tx {
	txIn, total := u.TransferTxIn(from, amount)
	if len(txIn) == 0 {
		panic("TransferTx: Insufficient balance")
	}
	txOut := []*TxOut{&TxOut{amount, to.PubKeyHash()}}
	change := total - amount
	if change > 0 {
		txOut = append(txOut, &TxOut{change, from.PubKeyHash()})
	}
	tx := &Tx{txIn, txOut}
	tx.Sign(from)
	return tx
}
