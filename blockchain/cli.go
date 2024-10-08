package blockchain

import (
	"fmt"
	"strconv"
)

func Wallet_(args []string) {
	if len(args) < 1 {
		fmt.Printf(
			"Usage:  blockchain wallet command args...\n\t" +
				"balance holder - get balance of holder wallet\n\t" +
				"create - create a new wallet\n\t" +
				"delete holder - delete wallet of holder\n\t" +
				"list - list all wallets\n",
		)
		return
	}
	method := args[0]
	db := GetDatabase()
	defer db.Close()
	ws := db.Wallets()
	if ws == nil {
		ws = new(Wallets)
		*ws = make(Wallets)
	}
	switch method {
	case "balance":
		holder := args[1]
		bc := db.Blockchain()
		wallet := ws.Wallet(holder)
		if wallet == nil {
			fmt.Println("Cli.Wallet: Failed to Get Wallet: Wallet does not exist")
			return
		}
		u := db.UTXOSet()
		if u == nil {
			u = new(UTXOSet)
			*u = make(UTXOSet)
			u.Index(bc)
			db.SetUTXOSet(u)
		}
		utxo := u.UnspentTxOuts(wallet)
		balance := 0
		for _, out := range utxo {
			balance += out.Value
		}
		fmt.Printf("Balance of %v: %v\n", wallet.Address(), balance)
	case "create":
		wallet := ws.NewWallet()
		db.SetWallets(ws)
		fmt.Println(wallet.Address())
	case "list":
		var wallets []string
		for wallet := range *ws {
			wallets = append(wallets, wallet)
		}
		fmt.Println(wallets)
	case "delete":
		holder := args[1]
		ws.Delete(holder)
		db.SetWallets(ws)
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
	db := GetDatabase()
	defer db.Close()
	ws := db.Wallets()
	if ws == nil {
		ws = new(Wallets)
		*ws = make(Wallets)
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
	bc := db.Blockchain()
	u := db.UTXOSet()
	if u == nil {
		u = new(UTXOSet)
		*u = make(UTXOSet)
		u.Index(bc)
		db.SetUTXOSet(u)
	}
	bc.Send(sender, receiver, amount, u)
}

func Mine(args []string) {
	if len(args) < 1 {
		fmt.Printf("Usage: blockchain mine miner - mine transactions from pool\n")
		return
	}
	miner := args[0]
	db := GetDatabase()
	defer db.Close()
	ws := db.Wallets()
	if ws == nil {
		ws = new(Wallets)
		*ws = make(Wallets)
	}
	bc := db.Blockchain()
	wallet := ws.Wallet(miner)
	if wallet == nil {
		fmt.Println("Cli.Mine: Failed to Get Wallet: Wallet does not exist")
		return
	}
	u := db.UTXOSet()
	if u == nil {
		u = new(UTXOSet)
		*u = make(UTXOSet)
		u.Index(bc)
		db.SetUTXOSet(u)
	}
	bc.Mine(wallet, u)
}

func Verify() {
	db := GetDatabase()
	defer db.Close()
	bc := db.Blockchain()
	fmt.Printf("Valid: %v\n", bc.Verify())
}

func Print() {
	db := GetDatabase()
	defer db.Close()
	bc := db.Blockchain()
	bc.Print()
}
