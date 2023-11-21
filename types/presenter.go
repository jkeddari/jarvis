package types

type ServerInfo struct {
	Infos []BlockchainInfo `json:"blockchains"`
}

type BlockchainInfo struct {
	Name    Blockchain `json:"name"`
	Symbols []Symbol   `json:"symbols"`
}

type BlockchainsInfo []BlockchainInfo

type APIError struct {
	Error string `json:"error"`
}

type Balance struct {
	Amount float64 `json:"amount"`
	Symbol Symbol  `json:"symbol"`
}
