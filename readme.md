# A simple blockchain implementation in Golang

This blockchain implementation had been written using [tutorial by jeiwan](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) with minor changes in implementation details. Some types had been changed from byte slice to hexadecimal formatted string for the ease of work and implementation. Some names had been shorten as excessively mouthful to type and reason about. Some obscure structures, like CLI, had been eliminated, as unnecessary. This version implements all core features except ones from the last part. Below you can see a brief description of each file.

| File Name | File Description |
|-----------|------------------|
| base58.go | Implements [Base58](https://en.wikipedia.org/wiki/Binary-to-text_encoding#Base58) encoding |
| block.go  | Implements the basic structures and methods of blocks, fundamental part of the blockchain |
| blockchain.go | Implements the basic structures and methods of blockchain, core technology of this module |
| iterator.go | Implements [Iterator](https://en.wikipedia.org/wiki/Iterator) of blockchain structure, allowing to read blocks one by one rather than loading everything at once |
| main.go | Entry point of the module, contains CLI processing using [Flag](https://pkg.go.dev/flag) package |
| transactions.go | Implements Transaction, which stores assets transfer between wallets |
| utxoset.go | Implements [UTXO Set](https://en.wikipedia.org/wiki/Unspent_transaction_output#UTXO_set), optimization of transaction processing, allowing to store unspent transaction outputs instead of full processing of the entire chain each time the output is needed |
| wallet.go | Implements [Cryptocurrency Wallet](https://en.wikipedia.org/wiki/Cryptocurrency_wallet), a mechanism for identifying owner of a particular asset using open key cryptography algorithms |