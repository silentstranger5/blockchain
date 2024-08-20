package blockchain

type Record struct {
	Wallet  *Wallet
	Balance int64
}

type Records struct {
	Records []*Record
}
