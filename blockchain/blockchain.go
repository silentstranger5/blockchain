package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

const difficulty = 16
const reward = 10

type Blockchain struct {
	Blocks     []*Block
	Pool       Txs
	Difficulty int
	Valid      bool
}

func (bc *Blockchain) LastHash() []byte {
	var lastHash []byte
	if len(bc.Blocks) > 0 {
		lastHash = bc.Blocks[0].Header.Hash
	}
	return lastHash
}

func (bc *Blockchain) CoinBaseTx(wallet *Wallet) {
	txin := []*TxIn{&TxIn{}}
	txout := []*TxOut{&TxOut{reward, wallet.PubKeyHash()}}
	tx := &Tx{txin, txout}
	tx.Sign(wallet)
	bc.Pool = append(bc.Pool, tx)
}

func (bc *Blockchain) TransferTx(from, to *Wallet, amount int) error {
	txIn, total := bc.TransferTxIn(from, amount)
	if len(txIn) == 0 {
		return fmt.Errorf("Insufficient balance")
	}
	txOut := []*TxOut{&TxOut{amount, to.PubKeyHash()}}
	change := total - amount
	if change > 0 {
		txOut = append(txOut, &TxOut{change, from.PubKeyHash()})
	}
	tx := &Tx{txIn, txOut}
	tx.Sign(from)
	bc.Pool = append(Txs{tx}, bc.Pool...)
	return nil
}

func (bc *Blockchain) TransferTxIn(from *Wallet, amount int) ([]*TxIn, int) {
	spendable, total := bc.FindSTXO(from, amount)
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

func (bc *Blockchain) FindSTXO(wallet *Wallet, amount int) (map[string][]int, int) {
	spendable := make(map[string][]int)
	unspent := bc.FindUTX(wallet)
	total := 0

	for _, tx := range unspent {
		txHash := fmt.Sprintf("%x", tx.Hash())
		for idx, out := range tx.TxOut {
			if out.LockedWith(wallet) {
				total += out.Value
				spendable[txHash] = append(spendable[txHash], idx)
			}
			if total >= amount {
				break
			}
		}
	}
	return spendable, total
}

func (bc *Blockchain) FindUTX(wallet *Wallet) []*Tx {
	spent := make(map[string][]int)
	unspent := make([]*Tx, 0)
	for _, tx := range bc.Pool {
		txHashStr := fmt.Sprintf("%x", tx.Hash())
		for _, in := range tx.TxIn {
			if tx.SignedWith(wallet) {
				txOutHashStr := fmt.Sprintf("%x", in.TxOutHash)
				spent[txOutHashStr] = append(spent[txOutHashStr], in.TxOutIndex)
			}
		}
		for outIdx, out := range tx.TxOut {
			spentout := false
			for _, idx := range spent[txHashStr] {
				if idx == outIdx {
					spentout = true
					break
				}
			}
			if spentout {
				continue
			}
			if out.LockedWith(wallet) {
				unspent = append(unspent, tx)
			}
		}
	}
	for _, block := range bc.Blocks {
		for _, tx := range block.Txs {
			txHashStr := fmt.Sprintf("%x", tx.Hash())
			for _, in := range tx.TxIn {
				if tx.SignedWith(wallet) {
					txOutHashStr := fmt.Sprintf("%x", in.TxOutHash)
					spent[txOutHashStr] = append(spent[txOutHashStr], in.TxOutIndex)
				}
			}
			for outIdx, out := range tx.TxOut {
				spentout := false
				for _, idx := range spent[txHashStr] {
					if idx == outIdx {
						spentout = true
						break
					}
				}
				if spentout {
					continue
				}
				if out.LockedWith(wallet) {
					unspent = append(unspent, tx)
				}
			}
		}
	}
	return unspent
}

func (bc *Blockchain) FindUTXO(wallet *Wallet) []*TxOut {
	unspent := bc.FindUTX(wallet)
	utxo := make([]*TxOut, 0)
	for _, tx := range unspent {
		for _, out := range tx.TxOut {
			if out.LockedWith(wallet) {
				utxo = append(utxo, out)
			}
		}
	}
	return utxo
}

func (bc *Blockchain) MineBlock(miner *Wallet) error {
	bc.CoinBaseTx(miner)
	prevHash := bc.LastHash()
	header := NewBlockHeader(prevHash)
	txs := bc.Pool
	bc.Pool = nil
	block := &Block{header, txs}
	block = block.Mine(bc.Difficulty)
	bc.Blocks = append([]*Block{block}, bc.Blocks...)
	bc.Verify()
	return nil
}

func (bc *Blockchain) Verify() bool {
	result := true
	for n, block := range bc.Blocks {
		result = result && block.Verify()
		if n < len(bc.Blocks)-1 {
			result = result &&
				reflect.DeepEqual(
					block.Header.PrevHash,
					bc.Blocks[n+1].Header.Hash,
				)
		}
		if !result {
			break
		}
	}
	bc.Valid = result
	return result
}

func GetBlockchain() (*Blockchain, error) {
	_, err := os.Stat("data")
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("data", 0750)
		if err != nil {
			return nil, err
		}
	}
	_, err = os.Stat("data/blockchain.json")
	if errors.Is(err, os.ErrNotExist) {
		bc := &Blockchain{}
		bc.Difficulty = difficulty
		data, err := json.Marshal(bc)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile("data/blockchain.json", data, 0666)
		if err != nil {
			return nil, err
		}
		return bc, nil
	}
	data, err := os.ReadFile("data/blockchain.json")
	if err != nil {
		return nil, err
	}
	bc := &Blockchain{}
	err = json.Unmarshal(data, bc)
	if err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc *Blockchain) Write() error {
	data, err := json.Marshal(bc)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/blockchain.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}
