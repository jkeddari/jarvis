package types

type (
	Blockchain string
	Owner      string
	Symbol     string
)

const (
	Bitcoin  Blockchain = "bitcoin"
	Ethereum Blockchain = "ethereum"

	Binance  Owner = "binance"
	Coinbase Owner = "coinbase"
	Kraken   Owner = "kraken"

	ETH  Symbol = "eth"
	USDT Symbol = "usdt"
)

type Transaction struct {
	Blockchain  Blockchain `json:"blockchain"`
	BlockNumber uint64     `json:"number"`
	Hash        string     `json:"hash"`
	Timestamp   int64      `json:"timestamp"`
	Fee         float64    `json:"fee"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Amount      float64    `json:"amount"`
	Symbol      Symbol     `json:"symbol"`
}

type AddressOwner struct {
	Chain   Blockchain `json:"chain"`
	Address string     `json:"address"`
	Owner   Owner      `json:"owner"`
	Match   float32    `json:"match"`
}
