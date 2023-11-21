package client

import (
	"context"

	"github.com/jkeddari/jarvis/types"
	"github.com/rs/zerolog"
)

// Config
type Config struct {
	DBPath string
	URL    string
	Ctx    context.Context
	Logger *zerolog.Logger
}

type Client interface {
	Run()

	GetStatus() (*types.BlockchainStatus, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
	GetBalance(address string) (*types.Balance, error)
	GetTransactionByHash(hash string) (*types.Transaction, error)
	GetTransactionsFromNumber(number, limit uint64) (types.Transactions, error)
	GetTransactionsForAddress(address string) (types.Transactions, error)
}
