package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// printUsage prints a help message
func printUsage() {
	fmt.Printf("blockchain 	- cli/db blockchain tool\n" +
		"\tcreatewallet		- create a new wallet\n" +
		"\tlistwallets		- list the wallet addresses\n" +
		"\tcreateblockchain ADR	- create a blockchain\n" +
		"\tsend	FROM TO AMNT 	- transfer coins\n" +
		"\tprintchain		- print blockchain\n")
	os.Exit(1)
}

// main is an entry point of the module where CLI arguments
// are parsed and processed
func main() {
	if len(os.Args) < 2 {
		printUsage()
	}
	switch os.Args[1] {
	case "createblockchain":
		address := os.Args[2]
		if !ValidAddress(address) {
			log.Fatal("Invalid address")
		}
		CreateBlockchain(address)
		os.Exit(0)
	case "createwallet":
		wallets, err := GetWallets()
		if err != nil {
			log.Fatal(err)
		}
		address := wallets.CreateWallet()
		wallets.SaveToFile()
		fmt.Printf("Your new address: %s\n", address)
		os.Exit(0)
	case "listwallets":
		wallets, err := GetWallets()
		if err != nil {
			log.Fatal(err)
		}
		addresses := wallets.GetAddresses()
		for _, address := range addresses {
			fmt.Println(address)
		}
		os.Exit(0)
	}
	bc := GetBlockchain()
	defer bc.DB.Close()
	switch os.Args[1] {
	case "send":
		from, to := os.Args[2], os.Args[3]
		if !(ValidAddress(from) && ValidAddress(to)) {
			log.Fatal("Invalid address")
		}
		amount, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		bc.Send(from, to, amount)
	case "printchain":
		bc.PrintChain()
	case "getbalance":
		address := os.Args[2]
		if !ValidAddress(address) {
			log.Fatal("Invalid address")
		}
		bc.GetBalance(address)
	default:
		printUsage()
	}
}
