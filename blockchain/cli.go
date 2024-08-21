package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Wallet_(args []string) {
	if len(args) < 1 {
		fmt.Printf(
			"Usage:  blockchain wallet command args...\n\t" +
				"balance holder - get balance of holder wallet\n\t" +
				"create - create a new wallet\n\t" +
				"list - list all wallets\n\t" +
				"delete holder - delete wallet of holder\n",
		)
		return
	}
	method := args[0]
	ws, err := GetWallets()
	if err != nil {
		fmt.Printf("Cli.Wallet: Failed to Get Wallets: %v\n", err)
		return
	}
	defer ws.Write()
	switch method {
	case "balance":
		holder := args[1]
		bc, err := GetBlockchain()
		if err != nil {
			fmt.Printf("Cli.Wallet: Failed to GetBlockchain: %v\n", err)
			return
		}
		wallet := ws.Wallet(holder)
		if wallet == nil {
			fmt.Println("Cli.Wallet: Failed to Get Wallet: Wallet does not exist")
			return
		}
		fmt.Printf("Balance of %v: %v\n", wallet.Address(), bc.Balance(wallet))
	case "create":
		wallet := ws.NewWallet()
		fmt.Println(wallet.Address())
	case "list":
		wallets := make([]string, 0)
		for wallet := range *ws {
			wallets = append(wallets, wallet)
		}
		fmt.Println(wallets)
	case "delete":
		holder := args[1]
		ws.Delete(holder)
	}
}

func Send(args []string) {
	if len(args) < 3 {
		fmt.Printf(
			"Usage: blockchain send from to amount - record a transfer transaction " +
				"between wallets\n",
		)
		return
	}
	from, to := args[0], args[1]
	ws, err := GetWallets()
	if err != nil {
		fmt.Printf("Cli.Send: Failed to GetWallets: %v\n", err)
		return
	}
	bc, err := GetBlockchain()
	if err != nil {
		fmt.Printf("Cli.Send: Failed to GetBlockchain: %v\n", err)
		return
	}
	sender := ws.Wallet(from)
	if sender == nil {
		fmt.Println("Cli.Send: Failed to Get Wallet: Wallet does not exist")
		return
	}
	receiver := ws.Wallet(to)
	if receiver == nil {
		fmt.Println("Cli.Send: Failed to Get Wallet: Wallet does not exist")
		return
	}
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Cli.Send: Failed to Record TransferTx: Invalid Amount Value\n")
		return
	}
	bc.TransferTx(sender, receiver, amount)
	bc.Write()
}

func Mine(args []string) {
	if len(args) < 1 {
		fmt.Printf("Usage: blockchain mine miner - mine transactions from pool\n")
		return
	}
	miner := args[0]
	ws, err := GetWallets()
	if err != nil {
		fmt.Printf("Cli.Mine: Failed to GetWallets: %v\n", err)
		return
	}
	bc, err := GetBlockchain()
	if err != nil {
		fmt.Printf("Cli.Mine: Failed to GetBlockchain: %v\n", err)
		return
	}
	wallet := ws.Wallet(miner)
	if wallet == nil {
		fmt.Println("Cli.Mine: Failed to Get Wallet: Wallet does not exist")
		return
	}
	err = bc.MineBlock(wallet)
	if err != nil {
		fmt.Printf("Cli.Mine: Failed to MineBlock: %v\n", err)
		return
	}
	bc.Write()
}

func Verify() {
	bc, err := GetBlockchain()
	defer bc.Write()
	if err != nil {
		fmt.Printf("Cli.Print: Failed to GetBlockchain: %v\n", err)
		return
	}
	fmt.Printf("Valid: %v\n", bc.Verify())
}

func Print() {
	bc, err := GetBlockchain()
	if err != nil {
		fmt.Printf("Cli.Print: Failed to GetBlockchain: %v\n", err)
		return
	}
	data, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		fmt.Printf("Cli.Print: Failed to MarshalIndent: %v\n", err)
		return
	}
	fmt.Print(string(data))
}
