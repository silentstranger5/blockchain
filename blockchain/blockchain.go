package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const difficulty = 16
const reward = 10

type Blockchain struct {
	Blocks     []*Block
	Pool       Txs
	Difficulty int64
	Valid      bool
}

func (bc *Blockchain) PrevHash() Hash {
	var prevHash Hash
	if len(bc.Blocks) > 0 {
		prevHash = bc.Blocks[len(bc.Blocks)-1].Header.Hash
	}
	return prevHash
}

func (bc *Blockchain) CoinBaseTx(owner *Wallet) {
	bc.Pool = append(bc.Pool, &Tx{&Wallet{"CoinBase"}, owner, reward})
}

func (bc *Blockchain) TransferTx(from, to *Wallet, amount int64) error {
	if bc.Balance(from) < amount {
		return fmt.Errorf("Blockchain.TransferTx: Failed to send %d from %s: Insufficient balance\n",
			amount, from.Holder,
		)
	}
	bc.Pool = append(bc.Pool, &Tx{from, to, amount})
	return nil
}

func (bc *Blockchain) MineBlock(miner *Wallet) error {
	bc.CoinBaseTx(miner)
	prevHash := bc.PrevHash()
	header := NewBlockHeader(prevHash)
	txs := bc.Pool
	bc.Pool = nil
	block := &Block{header, txs}
	block, err := block.Mine(bc.Difficulty)
	if err != nil {
		return err
	}
	bc.Blocks = append(bc.Blocks, block)
	bc.Verify()
	return nil
}

func (bc *Blockchain) Balance(wallet *Wallet) int64 {
	var balance int64
	for _, block := range bc.Blocks {
		for _, tx := range block.Txs {
			if *tx.From == *wallet {
				balance -= tx.Amount
			} else if *tx.To == *wallet {
				balance += tx.Amount
			}
		}
	}
	return balance
}

func (bc *Blockchain) Verify() {
	result := true
	for n, block := range bc.Blocks {
		ok, err := block.Verify()
		if err != nil {
			fmt.Printf("Blockchain.Verify: %v\n", err)
			return
		}
		result = result && ok
		if n > 0 {
			result = result &&
				block.Header.PrevHash ==
					bc.Blocks[n-1].Header.Hash
		}
		if !result {
			break
		}
	}
	bc.Valid = result
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
