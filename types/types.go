package types

// BlockchainStatus defines blockchain current status, such as start block number
// and last block number available into database.
type BlockchainStatus struct {
	StartNumber uint64 `json:"start_number"`
	EndNumber   uint64 `json:"end_number"`
}

type BlockchainsStatus []BlockchainStatus

type Transaction struct {
	BlockNumber uint64  `json:"number"`
	Hash        string  `json:"hash"`
	Timestamp   int64   `json:"timestamp"`
	Fee         float64 `json:"fee"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Amount      float64 `json:"amount"`
}

type Transactions []Transaction

type Block struct {
	Number    uint64       `json:"number"`
	Hash      string       `json:"hash"`
	Timestamp uint64       `json:"timestamps"`
	TXS       Transactions `json:"transactions"`
}

type AddressOwner struct {
	Chain   Blockchain `json:"chain"`
	Address string     `json:"address"`
	Owner   Owner      `json:"owner"`
	Match   float32    `json:"match"`
}
