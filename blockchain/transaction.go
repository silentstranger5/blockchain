package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"reflect"
)

type Tx struct {
	TxIn  []*TxIn
	TxOut []*TxOut
}

type TxIn struct {
	TxOutHash  Bytes
	TxOutIndex int
	Signature  Bytes
	PubKey     Bytes
}

type TxOut struct {
	Value      int
	PubKeyHash Bytes
}

type Txs []*Tx

func (txs Txs) Bytes() []byte {
	data := make([][]byte, 0)
	for _, tx := range txs {
		data = append(data, tx.Bytes())
	}
	return bytes.Join(data, nil)
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

func (in *TxIn) Bytes() []byte {
	return bytes.Join([][]byte{
		in.TxOutHash,
		IntToBytes(in.TxOutIndex),
		in.Signature,
		in.PubKey,
	}, nil)
}

func (out *TxOut) Bytes() []byte {
	return append(IntToBytes(out.Value), out.PubKeyHash...)
}

func (tx *Tx) Sign(w *Wallet) {
	privateKey := (*ecdsa.PrivateKey)(w)
	publicKey := (*ecdsa.PublicKey)(&w.PublicKey)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, tx.Strip().Hash())
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

func (tx *Tx) Strip() *Tx {
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

func (out *TxOut) LockedWith(w *Wallet) bool {
	return reflect.DeepEqual([]byte(out.PubKeyHash), w.PubKeyHash())
}

func (tx *Tx) SignedWith(w *Wallet) bool {
	walletPubKey := (*ecdsa.PublicKey)(&w.PublicKey)
	ok := true
	for _, in := range tx.TxIn {
		ok = ok && ecdsa.VerifyASN1(walletPubKey, tx.Strip().Hash(), in.Signature)
	}
	return ok
}
