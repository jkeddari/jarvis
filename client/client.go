package client

import (
	"context"

	"github.com/jkeddari/jarvis/types"
	"github.com/rs/zerolog"
)

const defaultMaxConcurrents = 50

// Config represents the configuration structure for Client
type Config struct {
	DBPath           string
	URL              string
	StreamURL        string
	Ctx              context.Context
	ConcurrentNumber int
	Logger           *zerolog.Logger
}

type Client interface {
	// Run subscribe to new block on Ethereum blockchain.
	Run() error

	// GetStatus returns blockchain status.
	GetStatus() (*types.BlockchainStatus, error)

	// GetBlockByNumber returns block by given number.
	GetBlockByNumber(number uint64) (*types.Block, error)

	// GetBlockTransactions returns transactions by given block number.
	GetBlockTransactions(number uint64) (types.Transactions, error)

	// GetBalance returns address Ethereum balance.
	GetBalance(address string) (*types.Balance, error)

	// GetTransactionByHash returns transaction by given hash.
	GetTransactionByHash(hash string) (*types.Transaction, error)

	// GetTransactionsForAddress returns all transaction for given address.
	GetTransactionsForAddress(address string) (types.Transactions, error)

	// GetAddressOwner returns address owner with the given address.
	GetAddressOwner(address string) (*types.AddressOwner, error)

	// SetAddressOwner write new address owner into database.
	SetAddressOwner(address string, owner types.Owner, match float32) error
}
