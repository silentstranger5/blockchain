# Simple Blockchain Implementation in Golang

This is a simple Blockchain Implementation in Golang. 
Data is stored in JSON. Transactions are simple. 

| Module Name | Description |
|-------------|-------------|
| block.go | Block is a unit of blockchain. It stores a collection of transactions with metadata |
| blockchain.go | Blockchain is a core technology of this project. It cointains a multiple blocks linked with a hash field |
| cli.go | Command-Line Interface entry point of application with argument parsing |
| marshal.go | Custom JSON Marshaling of Hash field |
| transaction.go | Transaction is a record of asset transfer between wallets. An atomic record of block. |
| utils.go  | Integer to Bytes converter utility function |
| wallet.go | Wallet denotes an asset holder in system |

## How to build

```
git clone https://github.com/silentstranger5/blockchain.git
cd blockchain
# you can either launch project interactively
go run blockchain
# or compile it
go build blockchain
./blockchain
```
