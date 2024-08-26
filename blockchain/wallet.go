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

	"blockchain/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = 00
const cslen = 4

type Wallet ecdsa.PrivateKey

func (w *Wallet) Bytes() []byte {
	if w == nil {
		return nil
	}

	return bytes.Join([][]byte{
		w.X.Bytes(),
		w.Y.Bytes(),
		w.D.Bytes(),
	}, []byte{})
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

func (w *Wallet) UnmarshalJSON(b []byte) error {
	type WalletAlias struct {
		elliptic.Curve `json:"omitempty"`
		X, Y, D        *big.Int
	}
	wa := new(WalletAlias)
	err := json.Unmarshal(b, wa)
	if err != nil {
		return err
	}
	wa.Curve = elliptic.P256()
	*w = Wallet{PublicKey: ecdsa.PublicKey{
		Curve: wa.Curve, X: wa.X, Y: wa.Y,
	}, D: wa.D}
	return nil
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
