package client

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/jkeddari/jarvis/db"
	"github.com/jkeddari/jarvis/types"
	"github.com/rs/zerolog"
)

const (
	defaultDBPath = "/tmp/jarvis_ethdb"
)

// ETHClient represents ethereum client with database.
type ETHClient struct {
	logger       zerolog.Logger
	db           db.DB
	client       *ethclient.Client
	clientMutex  sync.Mutex
	once         sync.Once
	ctx          context.Context
	url          string
	streamURL    string
	nbConcurrent int
}

// NewClient returns Ethereum client
func NewClient(config *Config) (Client, error) {
	if config.Ctx == nil {
		config.Ctx = context.Background()
	}

	c, err := ethclient.DialContext(config.Ctx, config.URL)
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
		client:       c,
		db:           db,
		ctx:          config.Ctx,
		url:          config.URL,
		streamURL:    config.StreamURL,
		nbConcurrent: defaultMaxConcurrents,
	}

	if config.Logger != nil {
		client.logger = *config.Logger
	}

	if config.ConcurrentNumber != 0 {
		client.nbConcurrent = config.ConcurrentNumber
	}

	return client, nil
}

func (c *ETHClient) Run() error {
	client, err := ethclient.DialContext(c.ctx, c.streamURL)
	if err != nil {
		return err
	}

	headers := make(chan *ethtypes.Header)
	sub, err := client.SubscribeNewHead(c.ctx, headers)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-sub.Err():
			return err
		case header := <-headers:
			block, err := client.BlockByHash(c.ctx, header.Hash())
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

func (c *ETHClient) GetStatus() (*types.BlockchainStatus, error) {
	return c.db.Status()
}

func (c *ETHClient) GetBlockByNumber(number uint64) (*types.Block, error) {
	if block, err := c.db.BlockByNumber(number); err == nil {
		return block, err
	}

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

func (c *ETHClient) GetBlockTransactions(number uint64) (types.Transactions, error) {
	return c.db.TransactionsForBlock(number)
}

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

func (c *ETHClient) GetTransactionsForAddress(address string) (types.Transactions, error) {
	if !common.IsHexAddress(address) {
		return nil, errors.New("bad ethereum address")
	}
	return c.db.TransactionsForAddress(address)
}

func (c *ETHClient) GetAddressOwner(address string) (*types.AddressOwner, error) {
	return c.db.GetAddressOwner(address)
}

func (c *ETHClient) SetAddressOwner(address string, owner types.Owner, match float32) error {
	addressOwner := types.AddressOwner{
		Chain:   types.Ethereum,
		Address: address,
		Owner:   owner,
		Match:   match,
	}
	return c.db.SetAddressOwner(addressOwner)
}

func (c *ETHClient) proccessBlock(ethBlock *ethtypes.Block) (*types.Block, error) {
	c.once.Do(func() {
		status := types.BlockchainStatus{
			StartNumber: ethBlock.NumberU64(),
			EndNumber:   ethBlock.NumberU64(),
		}
		c.db.SetStatus(status)
	})

	startTime := time.Now()

	b := types.Block{
		Number:             ethBlock.NumberU64(),
		Hash:               ethBlock.Hash().String(),
		Timestamp:          ethBlock.Time(),
		TransactionsNumber: len(ethBlock.Transactions()),
	}

	err := c.db.AddBlock(b)
	if err != nil {
		return nil, err
	}

	_, err = c.processTXs(ethBlock.Transactions()...)
	if err != nil {
		return nil, err
	}

	duration := time.Now().Sub(startTime)

	c.logger.Info().Any("block_number", b.Number).Any("process_duration", duration.String()).Msg("block added")

	return &b, c.db.UpdateNumber(b.Number)
}

func (c *ETHClient) processTX(ethTX *ethtypes.Transaction) (tx *types.Transaction, err error) {
	if ethTX == nil || ethTX.To() == nil {
		return nil, errors.New("contract creation")
	}

	client, err := ethclient.DialContext(c.ctx, c.url)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(c.ctx, ethTX.Hash())
	if err != nil {
		return nil, err
	}

	sender, err := client.TransactionSender(c.ctx, ethTX, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	fee := weiToEther(big.NewInt(int64(receipt.GasUsed * receipt.EffectiveGasPrice.Uint64())))

	tx = &types.Transaction{
		BlockNumber: receipt.BlockNumber.Uint64(),
		Timestamp:   ethTX.Time().Unix(),
		Hash:        ethTX.Hash().String(),
		Fee:         fee,
		From:        sender.String(),
		To:          ethTX.To().String(),
		Amount:      weiToEther(ethTX.Value()),
	}

	return tx, err
}

// processTXs convert ethereum transactions to types.Transaction, it using multiple ethclient concurrently.
func (c *ETHClient) processTXs(ethTXS ...*ethtypes.Transaction) (types.Transactions, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var txs types.Transactions

	semaphore := make(chan struct{}, c.nbConcurrent)

	for _, ethTX := range ethTXS {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(ethTX *ethtypes.Transaction, semaphore chan struct{}) {
			defer func() { <-semaphore }()
			defer wg.Done()

			tx, err := c.processTX(ethTX)
			if err != nil {
				c.logger.Error().Err(err)
				return
			}
			m.Lock()
			txs = append(txs, *tx)
			m.Unlock()
		}(ethTX, semaphore)
	}

	wg.Wait()
	return txs, c.db.AddTransactions(txs...)
}

func weiToEther(wei *big.Int) float64 {
	bfloat, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Float64()
	return bfloat
}
