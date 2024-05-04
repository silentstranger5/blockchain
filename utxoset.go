package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// utxobucket is a name of utxo set bucket
const utxobucket = "utxo"

// UTXOSet is a UTXO set structure
type UTXOSet struct {
	Blockchain *Blockchain
}

// Reindex updates a transaction outputs
func (u UTXOSet) Reindex() {
	db := u.Blockchain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(utxobucket))
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}
		_, err = tx.CreateBucket([]byte(utxobucket))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	unspent := u.Blockchain.FindUnspent()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxobucket))

		for txID, outs := range unspent {
			err = b.Put([]byte(txID), outs.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// FindSpendable returns spendable transaction outputs
func (u UTXOSet) FindSpendable(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspent := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxobucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := string(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.LockedWith(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspent[txID] = append(unspent[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return accumulated, unspent
}

// FindUnspent returns a slice of a transaction outputs
func (u UTXOSet) FindUnspent(pubKeyHash []byte) []TXOutput {
	var unspent []TXOutput
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxobucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.LockedWith(pubKeyHash) {
					unspent = append(unspent, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return unspent
}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxobucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get([]byte(vin.TXID))
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete([]byte(vin.TXID))
						if err != nil {
							return err
						}
					} else {
						err := b.Put([]byte(vin.TXID), updatedOuts.Serialize())
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)
			err := b.Put([]byte(tx.ID), newOutputs.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
