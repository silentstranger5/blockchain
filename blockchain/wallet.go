package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Wallet struct {
	Holder string
}

type Wallets []*Wallet

func NewWallets() Wallets {
	return Wallets{}
}

func (ws *Wallets) NewWallet(holder string) *Wallet {
	wallet := &Wallet{holder}
	*ws = append(*ws, wallet)
	return wallet
}

func (w *Wallet) Bytes() []byte {
	return []byte(w.Holder)
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

func (ws *Wallets) GetWallet(holder string) (*Wallet, error) {
	for _, wallet := range *ws {
		if wallet.Holder == holder {
			return wallet, nil
		}
	}
	return nil, fmt.Errorf("Wallet does not exist")
}

func (ws *Wallets) UpdateWallet(old, new_ string) {
	for _, wallet := range *ws {
		if wallet.Holder == old {
			wallet.Holder = new_
			break
		}
	}
}

func (ws *Wallets) DeleteWallet(holder string) {
	wind := -1
	for n, wallet := range *ws {
		if wallet.Holder == holder {
			wind = n
			break
		}
	}
	if wind > -1 {
		*ws = append((*ws)[:wind], (*ws)[wind+1:]...)
	}
}
