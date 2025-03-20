package sniffer

import (
	"encoding/json"
	"log/slog"

	"github.com/jkeddari/jarvis/internal/types"
	"github.com/nats-io/nats.go"
)

type Client interface {
	Run(chan *types.Transaction) error
	Stop() error
}

type sniffer struct {
	logger     *slog.Logger
	client     Client
	blockchain types.Blockchain
	txstream   chan *types.Transaction
	natsClient *nats.Conn
	subject    string
}

type Config struct {
	Blockchain types.Blockchain `env:"SNIFFER_BLOCKCHAIN"`
	URL        string           `env:"NATS_URL"`
	Subject    string           `env:"NATS_SUBJECT"`
}

func NewSniffer(c *Config, client Client, logger *slog.Logger) (*sniffer, error) {
	conn, err := nats.Connect(c.URL)
	if err != nil {
		return nil, err
	}

	return &sniffer{
		logger:     logger,
		client:     client,
		blockchain: c.Blockchain,
		txstream:   make(chan *types.Transaction),
		natsClient: conn,
		subject:    c.Subject,
	}, nil
}

func (s *sniffer) Run() error {
	go func() {
		for tx := range s.txstream {
			tx.Blockchain = s.blockchain
			payload, err := json.Marshal(tx)
			if err != nil {
				s.logger.Error("json", "Err", err)
			}
			err = s.natsClient.Publish(s.subject, payload)
			if err != nil {
				s.logger.Error("nats", "Err", err)
			}

		}
	}()
	return s.client.Run(s.txstream)
}

func (s *sniffer) Stop() error {
	s.logger.Info("stopping sniffer...")
	close(s.txstream)
	s.client.Stop()
	err := s.natsClient.Drain()
	s.natsClient.Close()

	return err
}
