package sniffer

import (
	"encoding/json"
	"log/slog"

	"github.com/jkeddari/jarvis/internal/types"
	"github.com/nats-io/nats.go"
)

type Client interface {
	Run(chan types.Transaction) error
}

type sniffer struct {
	logger     *slog.Logger
	client     Client
	blockchain types.Blockchain
	txstream   chan types.Transaction
	natsClient *nats.Conn
	subject    string
}

type Config struct {
	Blockchain types.Blockchain
	URL        string
	Subject    string
	Option     []nats.Option
}

func NewSniffer(c *Config, client Client, logger *slog.Logger) (*sniffer, error) {
	conn, err := nats.Connect(c.URL, c.Option...)
	if err != nil {
		return nil, err
	}

	return &sniffer{
		logger:     logger,
		client:     client,
		blockchain: c.Blockchain,
		txstream:   make(chan types.Transaction, 100),
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
		// TODO: Close this goroutine on Stop
	}()
	return s.client.Run(s.txstream)
}

func (s *sniffer) Stop() error {
	err := s.natsClient.Drain()
	s.natsClient.Close()
	return err
}
