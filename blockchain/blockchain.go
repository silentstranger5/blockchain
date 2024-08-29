package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

const difficulty = 16
const reward = 10

type Blockchain struct {
	Block      *Block `json:"-"`
	Pool       Txs
	Difficulty int
	Valid      bool
	DB         *Database `json:"-"`
}

func GetBlockchain(db *Database) *Blockchain {
	bc := &Blockchain{}
	db.Blockchain()
	bc.Block = db.NextBlock()
	pool := db.Pool()
	if pool != nil {
		bc.Pool = *pool
	}
	bc.Difficulty = difficulty
	bc.DB = db
	return bc
}

func (bc Blockchain) Send(from, to *Wallet, amount int, u *UTXOSet) {
	tx := TransferTx(from, to, amount, u)
	bc.Pool = append(Txs{tx}, bc.Pool...)
	u.Update(tx)
	bc.DB.SetPool(&bc.Pool)
	bc.DB.SetUTXOSet(u)
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
	bc.DB.AddBlock(block)
	bc.DB.SetPool(&bc.Pool)
	bc.DB.SetUTXOSet(u)
}

func (bc *Blockchain) TxByHash(txHash []byte) *Tx {
	bc.DB.Blockchain()
	block := bc.DB.NextBlock()
	for block != nil {
		for _, tx := range block.Txs {
			if reflect.DeepEqual(tx.Hash(), txHash) {
				return tx
			}
		}
		block = bc.DB.NextBlock()
	}
	return nil
}

func (bc *Blockchain) Verify() bool {
	result := true
	bc.DB.Blockchain()
	block := bc.DB.NextBlock()
	for block != nil {
		result = result && block.Verify()
		if next := bc.DB.PeekBlock(); next != nil {
			result = result &&
				reflect.DeepEqual(
					block.Header.PrevHash,
					next.Header.Hash,
				)
		}
		if !result {
			break
		}
		block = bc.DB.NextBlock()
	}
	bc.Valid = result
	return result
}

func (bc *Blockchain) LastHash() []byte {
	var lastHash []byte
	bc.DB.Blockchain()
	block := bc.DB.NextBlock()
	if block != nil {
		lastHash = block.Header.Hash
	}
	return lastHash
}

func (bc *Blockchain) UnspentTxOuts() map[string][]*TxOut {
	spent := make(map[string][]int)
	unspent := bc.Pool.UnspentTxOuts(spent)
	bc.DB.Blockchain()
	block := bc.DB.NextBlock()
	for block != nil {
		unspentBlock := block.Txs.UnspentTxOuts(spent)
		for k, v := range unspentBlock {
			unspent[k] = v
		}
		block = bc.DB.NextBlock()
	}
	return unspent
}

func (bc *Blockchain) Print() {
	bc.DB.Blockchain()
	block := bc.DB.NextBlock()
	for block != nil {
		data, err := json.MarshalIndent(block, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println("\"Block\": " + string(data))
		block = bc.DB.NextBlock()
	}
	bc.Verify()
	data, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println("\"Metadata\": " + string(data))
}

func (bc *Blockchain) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(bc)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func BlockchainDeserialize(data []byte) *Blockchain {
	bc := &Blockchain{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(bc)
	if err != nil {
		panic(err)
	}
	return bc
}

func GetBlockchain_() (*Blockchain, error) {
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
