package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

type BlockHeader struct {
	Timestamp int
	Nonce     int
	Hash      Hash
	PrevHash  Hash
}

type Hash []byte

func NewBlockHeader(prevHash []byte) BlockHeader {
	return BlockHeader{
		int(time.Now().Unix()),
		0,
		[]byte{},
		prevHash,
	}
}

func (b *Block) Hash() ([]byte, error) {
	bbytes, err := b.Bytes()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, err = h.Write(bbytes)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
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
		h.Hash,
		h.PrevHash,
	}, []byte{}), nil
}

func (b *Block) Mine(difficulty int) (*Block, error) {
	var err error
	var hash []byte
	var hashInt big.Int
	target := big.NewInt(1)
	target = target.Lsh(target, uint(256-difficulty))
	fmt.Println("Mining a New Block")
	for b.Header.Nonce < math.MaxInt {
		hash, err = b.Hash()
		if err != nil {
			return nil, err
		}
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
	return b, nil
}

func (b *Block) Verify() (bool, error) {
	bc := *b
	hash := b.Header.Hash
	bc.Header.Hash = []byte{}
	bchash, err := bc.Hash()
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(hash, bchash), nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h))
}

func (h *Hash) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*h, err = hex.DecodeString(s)
	if err != nil {
		return err
	}
	return nil
}
