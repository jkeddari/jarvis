package db

import "github.com/jkeddari/jarvis/types"

type dbBlock struct {
	Number    uint64
	Hash      string
	Timestamp uint64
}

type DB interface {
	// DropAll erase all database.
	DropAll() error

	// SetStatus write blockchain status into database.
	SetStatus(status types.BlockchainStatus) error

	// UpdateNumber update last block number.
	UpdateNumber(number uint64) error

	// AddBlock write new block into database.
	AddBlock(block types.Block) error

	// AddTransactions write new transactions into database.
	AddTransactions(txs ...types.Transaction) error

	// Status returns blockchain status.
	Status() (*types.BlockchainStatus, error)

	// BlockByNumber returns Block with the given number.
	BlockByNumber(number uint64) (*types.Block, error)

	// TransactionByHash returns transaction with the given hash.
	TransactionByHash(hash string) (*types.Transaction, error)

	// TransactionsFromBlock returns transactions from the given block number (default limit : 1000)
	TransactionsFromBlock(number, limit uint64) (types.Transactions, error)

	// TransactionsForAddress returns all transactions send by the given address.
	TransactionsForAddress(address string) (types.Transactions, error)
}