package blockchain

import (
	"errors"
	"os"

	bolt "go.etcd.io/bbolt"
)

type Database struct {
	DB  *bolt.DB
	Key []byte
}

const bcbucket = "blockchain"
const poolkey = "pool"
const tipkey = "tip"
const utxokey = "utxo"
const wskey = "wallets"
const dbpath = "data/blockchain.db"

func GetDatabase() *Database {
	_, err := os.Stat("data")
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("data", 0750)
		if err != nil {
			panic(err)
		}
	}
	db := &Database{}
	db.Open(dbpath)
	return db
}

func (d *Database) Open(path string) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		panic(err)
	}
	d.DB = db
}

func (d *Database) Close() {
	d.DB.Close()
}

func (d *Database) Blockchain() {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		var err error
		if b == nil {
			b, err = tx.CreateBucket([]byte(bcbucket))
			if err != nil {
				return err
			}
		}
		lastHash := b.Get([]byte(tipkey))
		d.Key = lastHash
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (d *Database) NextBlock() *Block {
	var data []byte
	err := d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		data = b.Get(d.Key)
		return nil
	})
	if err != nil {
		panic(err)
	}
	if data == nil {
		return nil
	}
	block := BlockDeserialize(data)
	d.Key = block.Header.PrevHash
	return block
}

func (d *Database) PeekBlock() *Block {
	key := d.Key
	block := d.NextBlock()
	d.Key = key
	return block
}

func (d *Database) AddBlock(block *Block) {
	hash := block.Header.Hash
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		err := b.Put(hash, block.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte(tipkey), hash)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (d *Database) Pool() *Txs {
	var data []byte
	err := d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		var err error
		if b == nil {
			b, err = tx.CreateBucket([]byte(bcbucket))
			if err != nil {
				return err
			}
		}
		data = b.Get([]byte(poolkey))
		return nil
	})
	if err != nil {
		panic(err)
	}
	if data == nil {
		return nil
	}
	return TxsDeserialize(data)
}

func (d *Database) SetPool(pool *Txs) {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		b.Put([]byte(poolkey), pool.Serialize())
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (d *Database) CleanPool() {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		b.Delete([]byte(poolkey))
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (d *Database) Wallets() *Wallets {
	var data []byte
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		var err error
		if b == nil {
			b, err = tx.CreateBucket([]byte(bcbucket))
			if err != nil {
				return err
			}
		}
		data = b.Get([]byte(wskey))
		return nil
	})
	if err != nil {
		panic(err)
	}
	if data == nil {
		return nil
	}
	return WalletsDeserialize(data)
}

func (d *Database) SetWallets(ws *Wallets) {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		b.Put([]byte(wskey), ws.Serialize())
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (d *Database) UTXOSet() *UTXOSet {
	var data []byte
	err := d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		var err error
		if b == nil {
			b, err = tx.CreateBucket([]byte(bcbucket))
			if err != nil {
				return err
			}
		}
		data = b.Get([]byte(utxokey))
		return nil
	})
	if err != nil {
		panic(err)
	}
	if data == nil {
		return nil
	}
	return UTXOSetDeserialize(data)
}

func (d *Database) SetUTXOSet(u *UTXOSet) {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcbucket))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		b.Put([]byte(utxokey), u.Serialize())
		return nil
	})
	if err != nil {
		panic(err)
	}
}
