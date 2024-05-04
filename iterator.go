package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator is an iterator of a blockchain
type BlockchainIterator struct {
	CurrentHash string
	DB          *bolt.DB
}

// Iterator constructs an iterator from the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.Tip, bc.DB}
}

// Next retrieves the next block in the blockchain
func (i *BlockchainIterator) Next() *Block {
	var data []byte
	err := i.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksbucket))
		data = b.Get([]byte(i.CurrentHash))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	block := Deserialize(data)
	i.CurrentHash = block.PrevHash
	return block
}
