# Simple Blockchain Implementation in Golang

This is a simple Blockchain Implementation in Golang. 

## Introduction

*Hashing algorithm* - produces a unique representation of data. If data is changed, output of the hashing 
algorithm differs drastically. In this module, most of the hashing is done by SHA256 algorithm.

*Block* - allows to store an arbitrary data. Each block contains a hash of the previous block, 
which "links" blocks together. Thus, the collection of blocks constitues a *chain*. Once you change the data 
in any of the blocks, this chain of blocks "breaks".

*Wallet* - a collection of public and private keys. *Private key* is used in order to produce a signature. 
Signature is used to transfer money, thus private key must be stored in secret. 
*Public key*, on the contrary, is used to verify signature and to confirm identity of spending wallet. 
It does not require discretion and can be shared freely.

*Transaction* - a record of asset transfer. Consists of inputs and outputs. *Outputs* store *value*, which 
can be spent. *Inputs* point to other outputs, and indicate a source of value in a transaction. All input value 
in a transaction must be spent. If there is any difference between input value and value intended for transfer, 
the difference is sent back to the original wallet. This difference is called *change*.

## Description

Data is stored as JSON for two reasons:

1. It is simple to work with
2. It is human readable, which allows to reuse code to print and debugging

Blocks are prepended onto blockchain in order to keep track of spent transaction outputs more easily.
Blockchain implements a transaction pool, which allows to follow the original model more accurately.
Signatures (with ECDSA) provide a source of entropy in transactions which allows to distinguish them.
Wallets are implemented as ECDSA private key defined by a set of parameters on elliptic curve.

| Module Name | Description |
|-------------|-------------|
| base58 | Base58 encoding implementation |
| block.go | Block is a unit of blockchain. It stores a collection of transactions with metadata |
| blockchain.go | Blockchain is a core technology of this project. It cointains a multiple blocks linked with a hash field |
| cli.go | Command-Line Interface entry point of application with argument parsing |
| transaction.go | Transaction is a record of asset transfer between wallets. An atomic record of block. |
| utils.go  | Integer to Bytes converter utility function |
| wallet.go | Wallet denotes an asset holder in system |

## How to build

```
git clone https://github.com/silentstranger5/blockchain.git
cd blockchain
# install dependencies
go get blockchain
# you can either launch project interactively
go run blockchain
# or compile it
go build blockchain
./blockchain
```
