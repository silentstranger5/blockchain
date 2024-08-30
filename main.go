package main

import (
	"blockchain/blockchain"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(
			"Usage:  blockchain command args...\n\t" +
				"wallet - manage wallets\n\t" +
				"mine - mine transactions from pool into block\n\t" +
				"print - print blockchain data\n\t" +
				"send - record a transfer transaction\n\t" +
				"verify - verify a blockchain integrity\n",
		)
		return
	}
	method := os.Args[1]
	args := os.Args[2:]
	switch method {
	case "wallet":
		blockchain.Wallet_(args)
	case "mine":
		blockchain.Mine(args)
	case "print":
		blockchain.Print()
	case "send":
		blockchain.Send(args)
	case "verify":
		blockchain.Verify()
	}
}
