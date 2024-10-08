package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"

	"blockchain/base58"

	"golang.org/x/crypto/ripemd160"
)

const (
	version = 00
	cslen   = 4
)

type Wallet ecdsa.PrivateKey

func (w *Wallet) Bytes() []byte {
	if w == nil {
		return nil
	}

	return bytes.Join([][]byte{
		w.X.Bytes(),
		w.Y.Bytes(),
		w.D.Bytes(),
	}, nil)
}

func (w *Wallet) Address() string {
	pubKeyHash := w.PubKeyHash()
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Encode(fullPayload)
	return string(address)
}

func (w *Wallet) PubKeyHash() []byte {
	h := sha256.New()
	pubKeyBytes := w.X.Bytes()
	h.Write(pubKeyBytes)
	shabytes := h.Sum(nil)
	r := ripemd160.New()
	r.Write(shabytes)
	ripebytes := r.Sum(nil)
	return ripebytes
}

func Checksum(payload []byte) []byte {
	h := sha256.New()
	h.Write(payload)
	first := h.Sum(nil)
	h.Reset()
	h.Write(first)
	second := h.Sum(nil)
	return second[:cslen]
}

func (w *Wallet) Serialize() []byte {
	pk := (*ecdsa.PrivateKey)(w)
	b, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		panic(err)
	}
	return b
}

func WalletDeserialize(data []byte) *Wallet {
	pk, err := x509.ParseECPrivateKey(data)
	if err != nil {
		panic(err)
	}
	return (*Wallet)(pk)
}

type Wallets map[string]*Wallet

func (ws *Wallets) NewWallet() *Wallet {
	curve := elliptic.P256()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	wallet := (*Wallet)(key)
	(*ws)[wallet.Address()] = wallet
	return wallet
}

func (ws *Wallets) Wallet(address string) *Wallet {
	return (*ws)[address]
}

func (ws *Wallets) Delete(address string) {
	delete(*ws, address)
}

// Sounds funny if you try to spell it
// but it actually stands for
// wallets serialized type
type wsst map[string][]byte

func (ws *Wallets) Serialize() []byte {
	wss := make(wsst)
	for k, v := range *ws {
		wss[k] = v.Serialize()
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(wss)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func WalletsDeserialize(data []byte) *Wallets {
	ws := &Wallets{}
	wss := make(wsst)
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&wss)
	if err != nil {
		panic(err)
	}
	for k, v := range wss {
		(*ws)[k] = WalletDeserialize(v)
	}
	return ws
}
