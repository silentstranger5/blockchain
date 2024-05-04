package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// genesis is a genesis block contents
const genesis = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// bits is an amount of first bits that need to be empty
const bits = 20

// Block is a block structure
type Block struct {
	Transactions []*Transaction
	Timestamp    int
	Hash         string
	PrevHash     string
	Nonce        int
	POW          bool
}

// GetHash obtains a hash string of a block
func (block *Block) GetHash() string {
	data := block.TxHash() +
		block.PrevHash +
		strconv.Itoa(block.Timestamp) +
		strconv.Itoa(block.Nonce)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// TxHash obtains a hash string of a block transactions
func (block *Block) TxHash() string {
	var txids string
	for _, tx := range block.Transactions {
		txids += tx.ID
	}
	hash := sha256.Sum256([]byte(txids))
	return fmt.Sprintf("%x", hash)
}

// NewBlock returns a new block constructed using
// collection of transactions and a previous block hash
func NewBlock(txs []*Transaction, prev string) *Block {
	block := Block{
		txs,
		int(time.Now().Unix()),
		"",
		prev,
		1,
		false,
	}
	fmt.Println("Mining a new block.")
	for !strings.HasPrefix(block.Hash, strings.Repeat("0", int(bits/4))) {
		block.Nonce++
		block.Hash = block.GetHash()
		fmt.Printf("\r%s", block.Hash)
	}
	block.POW = (block.GetHash() == block.Hash)
	fmt.Printf("\n\n")
	return &block
}

// NewGenesis returns a new genesis block
func NewGenesis(address string) *Block {
	tx := NewCoinbaseTX(address, genesis)
	return NewBlock([]*Transaction{tx}, "")
}

// NewTXBlock returns a new transaction block
func (bc *Blockchain) NewTXBlock(from, to string, amount int, u *UTXOSet) *Block {
	tx := bc.NewTransaction(from, to, amount, u)
	cbtx := NewCoinbaseTX(from, "")
	txs := []*Transaction{cbtx, tx}
	return NewBlock(txs, bc.Tip)
}

// Serialize returns a byte slice representation
// of a block
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(block)
	if err != nil {
		log.Fatal(err)
	}
	return buffer.Bytes()
}

// Deserialize returns a block structure
// obtained from serialized form
func Deserialize(data []byte) *Block {
	var block Block
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&block)
	if err != nil {
		log.Fatal(err)
	}
	return &block
}
