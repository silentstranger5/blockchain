package blockchain

import (
	"encoding/json"
	"errors"
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

func (bc *Blockchain) Send(from, to *Wallet, amount int, u *UTXOSet) {
	tx := TransferTx(from, to, amount, u)
	bc.Pool = append(Txs{tx}, bc.Pool...)
	u.Update(tx)
}

func (bc *Blockchain) Mine(miner *Wallet, u *UTXOSet) {
	tx := CoinBaseTx(miner)
	bc.Pool = append(Txs{tx}, bc.Pool...)
	u.Update(tx)
	prevHash := bc.LastHash()
	header := NewBlockHeader(prevHash)
	txs := bc.Pool
	bc.Pool = nil
	block := &Block{header, txs}
	block = block.Mine(bc.Difficulty)
	bc.Blocks = append([]*Block{block}, bc.Blocks...)
}

func (bc *Blockchain) TxByHash(txHash []byte) *Tx {
	for _, block := range bc.Blocks {
		for _, tx := range block.Txs {
			if reflect.DeepEqual(tx.Hash(), txHash) {
				return tx
			}
		}
	}
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

func (bc *Blockchain) LastHash() []byte {
	var lastHash []byte
	if len(bc.Blocks) > 0 {
		lastHash = bc.Blocks[0].Header.Hash
	}
	return lastHash
}

func (bc *Blockchain) UnspentTxOuts() map[string][]*TxOut {
	spent := make(map[string][]int)
	unspent := bc.Pool.UnspentTxOuts(spent)
	for _, block := range bc.Blocks {
		unspentBlock := block.Txs.UnspentTxOuts(spent)
		for k, v := range unspentBlock {
			unspent[k] = v
		}
	}
	return unspent
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
