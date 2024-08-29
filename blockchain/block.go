package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"time"
)

type Block struct {
	Header BlockHeader
	Txs    Txs
}

func (b *Block) Bytes() []byte {
	return append(
		b.Header.Bytes(),
		b.Txs.Bytes()...,
	)
}

func (b *Block) Hash() []byte {
	hash := sha256.Sum256(b.Bytes())
	return hash[:]
}

func (b *Block) Mine(difficulty int) *Block {
	var hash []byte
	var hashInt big.Int
	target := big.NewInt(1)
	target = target.Lsh(target, uint(256-difficulty))
	fmt.Println("Mining a New Block")
	for b.Header.Nonce < math.MaxInt {
		hash = b.Hash()
		hashInt.SetBytes(hash)
		fmt.Printf("\r%x", hash)
		if hashInt.Cmp(target) == -1 {
			break
		} else {
			b.Header.Nonce++
		}
	}
	b.Header.Hash = hash
	fmt.Println()
	return b
}

func (b *Block) Verify() bool {
	result := true
	bc := *b
	bc.Header.Hash = nil
	result = result && reflect.DeepEqual(
		[]byte(b.Header.Hash),
		bc.Hash(),
	)
	for _, tx := range b.Txs {
		for _, in := range tx.TxIn {
			pubKey := &ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     big.NewInt(0).SetBytes(in.PubKey[:32]),
				Y:     big.NewInt(0).SetBytes(in.PubKey[32:]),
			}
			result = result && ecdsa.VerifyASN1(
				pubKey, tx.Trim().Hash(), in.Signature,
			)
		}
	}
	return result
}

func (b *Block) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(b)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func BlockDeserialize(data []byte) *Block {
	b := &Block{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(b)
	if err != nil {
		panic(err)
	}
	return b
}

type BlockHeader struct {
	Timestamp int
	Nonce     int
	Hash      []byte
	PrevHash  []byte
}

func NewBlockHeader(prevHash []byte) BlockHeader {
	return BlockHeader{
		int(time.Now().Unix()),
		0,
		nil,
		prevHash,
	}
}

func (h *BlockHeader) Bytes() []byte {
	return bytes.Join([][]byte{
		IntToBytes(h.Timestamp),
		IntToBytes(h.Nonce),
		h.Hash,
		h.PrevHash,
	}, nil)
}
