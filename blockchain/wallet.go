package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = 00
const cslen = 4

type Wallet ecdsa.PrivateKey

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

func (w *Wallet) Address() string {
	pubKeyHash := PubKeyHash(w.PublicKey)
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Encode(fullPayload)
	return address
}

func PubKeyHash(pubKey ecdsa.PublicKey) []byte {
	h := sha256.New()
	pubKeyBytes := pubKey.X.Bytes()
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

func GetWallets() (*Wallets, error) {
	_, err := os.Stat("data")
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("data", 0750)
		if err != nil {
			return nil, err
		}
	}
	_, err = os.Stat("data/wallets.json")
	if errors.Is(err, os.ErrNotExist) {
		ws := &Wallets{}
		data, err := json.Marshal(ws)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile("data/wallets.json", data, 0666)
		if err != nil {
			return nil, err
		}
		return ws, nil
	}
	data, err := os.ReadFile("data/wallets.json")
	if err != nil {
		return nil, err
	}
	ws := &Wallets{}
	err = json.Unmarshal(data, ws)
	if err != nil {
		return nil, err
	}
	for _, wallet := range *ws {
		wallet.Curve = elliptic.P256()
	}
	return ws, nil
}

func (ws *Wallets) Write() error {
	data, err := json.Marshal(ws)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/wallets.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (ws *Wallets) Wallet(address string) *Wallet {
	return (*ws)[address]
}

func (ws *Wallets) Delete(address string) {
	delete(*ws, address)
}

func (w *Wallet) Bytes() []byte {
	if w == nil {
		return []byte{}
	}

	return bytes.Join([][]byte{
		w.X.Bytes(),
		w.Y.Bytes(),
		w.D.Bytes(),
	}, []byte{})
}

type MarshaledWallet struct {
	Curve any
	X     *big.Int
	Y     *big.Int
	D     *big.Int
}

func (w *Wallet) UnmarshalJSON(b []byte) error {
	mw := new(MarshaledWallet)
	err := json.Unmarshal(b, mw)
	if err != nil {
		return err
	}
	pubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: mw.X, Y: mw.Y}
	*w = Wallet{PublicKey: pubKey, D: mw.D}
	return nil
}
