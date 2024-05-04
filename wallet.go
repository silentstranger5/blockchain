package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"io"
	"log"
	"math/big"
	"os"

	"golang.org/x/crypto/ripemd160"
)

// cslen is a checksum length
const cslen = 4

// version is a byte representation of version
const version = byte(0x00)

// walletfile is a name of a wallets file
const walletfile = "wallets.dat"

// Wallet is a wallet structure
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Wallets is a structure containing
// a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallet returns a new wallet
func NewWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

// NewKeyPair returns a new key pair
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public
}

// GetAddress returns an address of a wallet
func (w Wallet) GetAddress() []byte {
	pubkeyhash := HashPubKey(w.PublicKey)
	payload := append([]byte{version}, pubkeyhash...)
	checksum := Checksum(payload)
	payload = append(payload, checksum...)
	address := Base58Encode(payload)
	return address
}

// HashPubKey returns a hash of a public key
func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(hash[:])
	if err != nil {
		log.Fatal(err)
	}
	return hasher.Sum(nil)
}

// Checksum returns a checksum of a payload hash
func Checksum(payload []byte) []byte {
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	return hash[:cslen]
}

// GetWallets returns a collection of wallets
func GetWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile()
	return &wallets, err
}

// CreateWallet creates a new wallet
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

// GetAddress returns an address of a wallet
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

// GetWallet returns a wallet from an address
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from a file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletfile); os.IsNotExist(err) {
		os.Create(walletfile)
	}
	fileContent, err := os.ReadFile(walletfile)
	if err != nil {
		return err
	}
	var wallets Wallets
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil && err != io.EOF {
		return err
	}
	if wallets.Wallets == nil {
		ws.Wallets = make(map[string]*Wallet)
	} else {
		ws.Wallets = wallets.Wallets
	}
	return nil
}

// SaveToFile writes wallets to a file
func (ws *Wallets) SaveToFile() {
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(walletfile, content.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// Code bellow contains implementation of a custom ECDSA serialization
type _PrivateKey struct {
	D          *big.Int
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (w *Wallet) GobEncode() ([]byte, error) {
	privKey := &_PrivateKey{
		D:          w.PrivateKey.D,
		PublicKeyX: w.PrivateKey.PublicKey.X,
		PublicKeyY: w.PrivateKey.PublicKey.Y,
	}

	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(privKey)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	var privKey _PrivateKey

	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&privKey)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{
		D: privKey.D,
		PublicKey: ecdsa.PublicKey{
			X:     privKey.PublicKeyX,
			Y:     privKey.PublicKeyY,
			Curve: elliptic.P256(),
		},
	}
	w.PublicKey = make([]byte, buf.Len())
	_, err = buf.Read(w.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func ValidAddress(address string) bool {
	if len(address) < 6 {
		return false
	}
	pubKeyHash := Base58Decode([]byte(address))
	actChecksum := pubKeyHash[len(pubKeyHash)-cslen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-cslen]
	tarChecksum := Checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Equal(actChecksum, tarChecksum)
}
