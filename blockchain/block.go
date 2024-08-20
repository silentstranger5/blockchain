package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"
)

type Block struct {
	Header BlockHeader
	Txs    Txs
}

type BlockHeader struct {
	Timestamp int64
	Nonce     int64
	Hash      Hash
	PrevHash  Hash
}

type Hash [32]byte

func NewBlockHeader(prevHash Hash) BlockHeader {
	return BlockHeader{
		time.Now().Unix(),
		0,
		Hash{},
		prevHash,
	}
}

func (b *Block) Hash() (Hash, error) {
	bbytes, err := b.Bytes()
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(bbytes), nil
}

func (b *Block) Bytes() ([]byte, error) {
	bhbytes, err := b.Header.Bytes()
	if err != nil {
		return nil, err
	}
	txbytes, err := b.Txs.Bytes()
	if err != nil {
		return nil, err
	}
	return append(bhbytes, txbytes...), nil
}

func (h *BlockHeader) Bytes() ([]byte, error) {
	tsbytes, err := IntToBytes(h.Timestamp)
	if err != nil {
		return nil, err
	}
	ncbytes, err := IntToBytes(h.Nonce)
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{
		tsbytes,
		ncbytes,
		h.Hash[:],
		h.PrevHash[:],
	}, []byte{}), nil
}

func (b *Block) Mine(difficulty int64) (*Block, error) {
	var err error
	var hash Hash
	var hashInt big.Int
	target := big.NewInt(1)
	target = target.Lsh(target, uint(256-difficulty))
	fmt.Println("Mining a New Block")
	for b.Header.Nonce < math.MaxInt {
		hash, err = b.Hash()
		if err != nil {
			return nil, err
		}
		hashInt.SetBytes(hash[:])
		fmt.Printf("\r%x", hash)
		if hashInt.Cmp(target) == -1 {
			break
		} else {
			b.Header.Nonce++
		}
	}
	b.Header.Hash = hash
	fmt.Println()
	return b, nil
}

func (b *Block) Verify() (bool, error) {
	bc := *b
	hash := b.Header.Hash
	bc.Header.Hash = Hash{}
	bchash, err := bc.Hash()
	if err != nil {
		return false, err
	}
	return bchash == hash, nil
}
