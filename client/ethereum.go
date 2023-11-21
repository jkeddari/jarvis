package client

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/jkeddari/jarvis/db"
	"github.com/jkeddari/jarvis/types"
	"github.com/rs/zerolog"
)

const defaultDBPath = "/tmp/jarvis_ethdb"

// ETHClient represents ethereum client with database.
type ETHClient struct {
	logger      zerolog.Logger
	db          db.DB
	client      *ethclient.Client
	clientMutex sync.Mutex
	once        sync.Once
	ctx         context.Context
}

// NewClient returns Ethereum client
func NewClient(config *Config) (Client, error) {
	c, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	path := defaultDBPath
	if config.DBPath != "" {
		path = config.DBPath
	}

	db, err := db.NewBadgerDB(path)
	if err != nil {
		return nil, err
	}

	client := &ETHClient{
		client: c,
		db:     db,
		ctx:    context.Background(),
	}

	if config.Ctx != nil {
		client.ctx = config.Ctx
	}

	if config.Logger != nil {
		client.logger = *config.Logger
	}

	return client, nil
}

func (c *ETHClient) GetStatus() (*types.BlockchainStatus, error) {
	return c.db.Status()
}

// GetBlockByNumber returns block by given number.
func (c *ETHClient) GetBlockByNumber(number uint64) (*types.Block, error) {
	if block, err := c.db.BlockByNumber(number); err == nil {
		return block, err
	}
	// FIXME : if block is not in db, code below is not called.
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	block, err := c.client.BlockByNumber(c.ctx, big.NewInt(int64(number)))
	if err != nil {
		return nil, err
	}
	b, err := c.proccessBlock(block)
	if err != nil {
		return nil, err
	}

	return b, err
}

// GetBalance returns address Ethereum balance
func (c *ETHClient) GetBalance(address string) (*types.Balance, error) {
	if !common.IsHexAddress(address) {
		return nil, errors.New("bad ethereum address")
	}

	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	number, err := c.client.BlockNumber(c.ctx)
	if err != nil {
		return nil, err
	}

	weiAmount, err := c.client.BalanceAt(c.ctx, common.HexToAddress(address), big.NewInt(int64(number)))
	return &types.Balance{
		Amount: weiToEther(weiAmount),
		Symbol: types.ETH,
	}, nil
}

// GetTransactionByHash returns transaction by given hash.
func (c *ETHClient) GetTransactionByHash(hash string) (*types.Transaction, error) {
	if tx, err := c.db.TransactionByHash(hash); err == nil {
		return tx, err
	}

	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	ethTX, pending, err := c.client.TransactionByHash(c.ctx, common.HexToHash(hash))
	if err != nil {
		return nil, err
	}

	if pending {
		return nil, errors.New("tx pending")
	}

	tx, err := c.processTX(ethTX)
	if err != nil {
		return nil, err
	}

	return tx, c.db.AddTransactions(*tx)
}

func (c *ETHClient) GetTransactionsFromNumber(number, limit uint64) (types.Transactions, error) {
	// TODO : get block if not exist on db
	return c.db.TransactionsFromBlock(number, limit)
}

func (c *ETHClient) GetTransactionsForAddress(address string) (types.Transactions, error) {
	return c.db.TransactionsForAddress(address)
}

// Run subscribe to new block on Ethereum blockchain.
func (c *ETHClient) Run() {
	headers := make(chan *ethtypes.Header)
	c.clientMutex.Lock()
	sub, err := c.client.SubscribeNewHead(c.ctx, headers)
	c.clientMutex.Unlock()
	if err != nil {
		c.logger.Fatal().Err(err)
	}

	for {
		select {
		case err := <-sub.Err():
			c.logger.Error().Err(err)
		case header := <-headers:
			c.clientMutex.Lock()
			block, err := c.client.BlockByHash(c.ctx, header.Hash())
			c.clientMutex.Unlock()
			if err != nil {
				c.logger.Error().Err(err)
				continue
			}
			if _, err := c.proccessBlock(block); err != nil {
				c.logger.Error().Err(err)
			}
		}
	}
}

func (c *ETHClient) processTX(ethTX *ethtypes.Transaction) (*types.Transaction, error) {
	if ethTX == nil || ethTX.To() == nil {
		return nil, errors.New("contract creation")
	}

	receipt, err := c.client.TransactionReceipt(c.ctx, ethTX.Hash())
	if err != nil {
		return nil, err
	}

	sender, err := c.client.TransactionSender(c.ctx, ethTX, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	// FIXME : bad fee returned
	fee := weiToEther(big.NewInt(int64(receipt.GasUsed * ethTX.GasPrice().Uint64())))

	return &types.Transaction{
		BlockNumber: receipt.BlockNumber.Uint64(),
		Timestamp:   ethTX.Time().Unix(),
		Hash:        ethTX.Hash().String(),
		Fee:         fee,
		From:        sender.String(),
		To:          ethTX.To().String(),
		Amount:      weiToEther(ethTX.Value()),
	}, nil
}

func (c *ETHClient) processTXs(ethTXS ...*ethtypes.Transaction) error {
	var txs types.Transactions
	for _, ethTX := range ethTXS {

		tx, err := c.processTX(ethTX)
		if err != nil {
			c.logger.Error().Err(err)
			continue
		}

		txs = append(txs, *tx)
	}

	return c.db.AddTransactions(txs...)
}

func (c *ETHClient) proccessBlock(ethBlock *ethtypes.Block) (*types.Block, error) {
	c.once.Do(func() {
		status := types.BlockchainStatus{
			StartNumber: ethBlock.NumberU64(),
			EndNumber:   ethBlock.NumberU64(),
		}
		c.db.SetStatus(status)
	})

	b := types.Block{
		Number:    ethBlock.NumberU64(),
		Hash:      ethBlock.Hash().String(),
		Timestamp: ethBlock.Time(),
		TXS:       types.Transactions{},
	}
	c.logger.Info().Any("block_number", b.Number).Msg("addblock")

	err := c.db.AddBlock(b)
	if err != nil {
		return nil, err
	}

	err = c.processTXs(ethBlock.Transactions()...)
	if err != nil {
		return nil, err
	}

	return &b, c.db.UpdateNumber(b.Number)
}

func weiToEther(wei *big.Int) float64 {
	bfloat, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Float64()
	return bfloat
}
