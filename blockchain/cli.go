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
				"create holder - create a new wallet for holder\n\t" +
				"list - list all wallets\n\t" +
				"update old new - update wallet holder from old to new\n\t" +
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
		wallet, err := ws.GetWallet(holder)
		if err != nil {
			fmt.Printf("Cli.Wallet: Failed to GetWallet: %v\n", err)
			return
		}
		fmt.Printf("Balance of %v: %v\n", wallet.Holder, bc.Balance(wallet))
	case "create":
		holder := args[1]
		ws.NewWallet(holder)
	case "list":
		data, err := json.MarshalIndent(ws, "", "  ")
		if err != nil {
			fmt.Printf("Cli.Wallet: Failed to Marshal JSON: %v\n", err)
			return
		}
		fmt.Printf(string(data))
	case "update":
		old := args[1]
		new_ := args[2]
		ws.UpdateWallet(old, new_)
	case "delete":
		holder := args[1]
		ws.DeleteWallet(holder)
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
	sender, err := ws.GetWallet(from)
	if err != nil {
		fmt.Printf("Cli.Send: Failed to GetWallet: %v\n", err)
		return
	}
	receiver, err := ws.GetWallet(to)
	if err != nil {
		fmt.Printf("Cli.Send: Failed to GetWallet: %v\n", err)
		return
	}
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Cli.Send: Failed to Record TransferTx: Invalid Amount Value\n")
		return
	}
	bc.TransferTx(sender, receiver, int64(amount))
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
	wallet, err := ws.GetWallet(miner)
	if err != nil {
		fmt.Printf("Cli.Mine: Failed to GetWallet: %v\n", err)
		return
	}
	err = bc.MineBlock(wallet)
	if err != nil {
		fmt.Printf("Cli.Mine: Failed to MineBlock: %v\n", err)
		return
	}
	bc.Write()
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
