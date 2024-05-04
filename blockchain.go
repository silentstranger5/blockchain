package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// blocksfile is name of blockchain database file
const blocksfile = "blockchain.db"

// blocksbucket is name of blocks bucket
const blocksbucket = "blocks"

// tipkey is name of tip key
const tipkey = "tip"

type Blockchain struct {
	DB  *bolt.DB
	Tip string
}

// dbExists checks if the database file exists in the module directory
func dbExists() bool {
	if _, err := os.Stat(blocksfile); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateBlockchain creates a new blockchain
func CreateBlockchain(address string) {
	if dbExists() {
		log.Fatal("Blockchain already exists.")
	}
	genesis := NewGenesis(address)
	db, err := bolt.Open(blocksfile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Creating a new blockchain.")
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksbucket))
		if err != nil {
			return err
		}
		err = b.Put([]byte(genesis.Hash), genesis.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte(tipkey), []byte(genesis.Hash))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success!")
	bc := &Blockchain{db, genesis.Hash}
	u := UTXOSet{bc}
	u.Reindex()
}

// GetBlockchain obtains blockchain
// from the database file
func GetBlockchain() *Blockchain {
	var db *bolt.DB
	var tip string
	if !dbExists() {
		log.Fatal("Blockchain does not exist. Try 'create'.")
	}
	db, err := bolt.Open(blocksfile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksbucket))
		tip = string(b.Get([]byte(tipkey)))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return &Blockchain{db, tip}
}

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	for _, tx := range block.Transactions {
		if !bc.VerifyTransaction(tx) {
			log.Fatal("Invalid Transaction.")
		}
	}
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksbucket))
		err := b.Put([]byte(block.Hash), block.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte(tipkey), []byte(block.Hash))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	bc.Tip = block.Hash
	fmt.Println("Success!")
}

// GetBalance retrieves an asset balance
// of the specified address
func (bc *Blockchain) GetBalance(address string) {
	var balance int
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	var u = UTXOSet{bc}
	unspent := u.FindUnspent(pubKeyHash)
	for _, out := range unspent {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// Send sends amount of assets between wallets
func (bc *Blockchain) Send(from, to string, amount int) {
	u := UTXOSet{bc}
	block := bc.NewTXBlock(from, to, amount, &u)
	bc.AddBlock(block)
	u.Update(block)
}

// PrintChain prints a string representation of the blockchain
func (bc *Blockchain) PrintChain() {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("%#v\n\n", *block)
		for _, tx := range block.Transactions {
			fmt.Printf("%#v\n\n", *tx)
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	fmt.Println("Chain Integrity: ", bc.Verify())
}

// Verify checks integrity of the blockchain
func (bc *Blockchain) Verify() bool {
	bci := bc.Iterator()
	var prev string
	for i := 0; ; i++ {
		block := bci.Next()
		if block.Hash != block.GetHash() ||
			i > 0 && block.Hash != prev {
			return false
		}
		prev = block.PrevHash
		if len(block.PrevHash) == 0 {
			return true
		}
	}
}

// FindTransaction retrieves a transaction
// with the specified ID from the blockchain
func (bc *Blockchain) FindTransaction(ID string) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if tx.ID == ID {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction signs the specified transaction
// with ECDSA private key
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TXID)
		if err != nil {
			log.Fatal(err)
		}
		prevTXs[prevTX.ID] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction checks validity of a signature of the transaction
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TXID)
		if err != nil {
			log.Fatal(err)
		}
		prevTXs[prevTX.ID] = prevTX
	}

	return tx.Verify(prevTXs)
}
