package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"math/big"
	"os"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/jkeddari/jarvis/internal/sniffer"
	"github.com/jkeddari/jarvis/internal/types"
	"github.com/joho/godotenv"
)

type ETHConfig struct {
	URL           string
	MinimumAmount float64
}

type ethClient struct {
	logger        *slog.Logger
	minimumAmount float64
	client        ethclient.Client
	headersChan   chan *ethtypes.Header
}

func newETHClient(config *ETHConfig, logger *slog.Logger) (sniffer.Client, error) {
	c, err := ethclient.DialContext(context.Background(), config.URL)
	if err != nil {
		return nil, err
	}

	return &ethClient{
		logger:        logger,
		client:        *c,
		headersChan:   make(chan *ethtypes.Header),
		minimumAmount: config.MinimumAmount,
	}, nil
}

func (c *ethClient) Run(txs chan types.Transaction) error {
	err := c.subscribeBlock(txs)
	if err != nil {
		return err
	}

	return nil
}

func (c *ethClient) subscribeBlock(txs chan types.Transaction) error {
	sub, err := c.client.SubscribeNewHead(context.Background(), c.headersChan)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-sub.Err():
			return err
		case header := <-c.headersChan:
			block, err := c.client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				c.logger.Error("block receive", "err", err)
				continue
			}
			if block != nil {
				c.logger.Info("block info", "number", block.NumberU64(), "txs_size", block.Transactions().Len())
				c.processTXs(txs, block.Transactions()...)
			}
		}
	}
}

func (c *ethClient) processTX(ethTX *ethtypes.Transaction) (tx *types.Transaction, err error) {
	if ethTX == nil || ethTX.To() == nil {
		return nil, errors.New("contract creation")
	}

	amount := weiToEther(ethTX.Value())
	if amount < 1 {
		return nil, nil
	}

	receipt, err := c.client.TransactionReceipt(context.Background(), ethTX.Hash())
	if err != nil {
		return nil, err
	}

	sender, err := c.client.TransactionSender(context.Background(), ethTX, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	fee := weiToEther(big.NewInt(int64(receipt.GasUsed * receipt.EffectiveGasPrice.Uint64())))

	return &types.Transaction{
		BlockNumber: receipt.BlockNumber.Uint64(),
		Hash:        ethTX.Hash().String(),
		Timestamp:   ethTX.Time().Unix(),
		Fee:         fee,
		From:        sender.String(),
		To:          ethTX.To().String(),
		Amount:      weiToEther(ethTX.Value()),
		Symbol:      "ETH",
	}, nil
}

func (c *ethClient) processTXs(txs chan types.Transaction, ethTXS ...*ethtypes.Transaction) error {
	for _, ethTX := range ethTXS {
		tx, err := c.processTX(ethTX)
		if err != nil {
			c.logger.Error("process tx", "err", err)
		}
		if tx != nil {
			txs <- *tx
		}
	}
	return nil
}

func weiToEther(wei *big.Int) float64 {
	bfloat, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Float64()
	return bfloat
}

func main() {
	logger := slog.Default()
	logger.Info("starting ethereum sniffer...")

	godotenv.Load()
	url := os.Getenv("RPC_URL")
	if url == "" {
		log.Fatal("no url found")
	}

	clientConfig := &ETHConfig{
		URL: url,
	}

	client, err := newETHClient(clientConfig, logger)
	if err != nil {
		logger.Error("new ethclient", "err", err)
		os.Exit(1)
	}

	sc := &sniffer.Config{
		Blockchain: "ethereum",
		URL:        "nats://192.168.117.2:4222",
		Subject:    "ethereum",
	}

	s, err := sniffer.NewSniffer(sc, client, logger)
	if err != nil {
		logger.Error("new sniffer", "err", err)
		os.Exit(1)
	}

	s.Run()
}
