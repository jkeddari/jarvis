package api

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jkeddari/jarvis/client"
	"github.com/rs/zerolog"
)

type Config struct {
	ETHConfig client.Config
}

type Server struct {
	Logger    zerolog.Logger
	router    *mux.Router
	ethClient client.Client
	once      sync.Once
}

func NewServer(conf Config) (*Server, error) {
	ethClient, err := client.NewClient(&conf.ETHConfig)
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	s := &Server{
		Logger:    logger,
		router:    mux.NewRouter(),
		ethClient: ethClient,
	}

	s.router.Methods("GET").Path("/status").HandlerFunc(s.handlerStatus())
	s.router.Methods("GET").Path("/{blockchain}/status").HandlerFunc(s.handlerBlockchainStatus())
	s.router.Methods("GET").Path("/{blockchain}/transaction/{hash}").HandlerFunc(s.handlerTransactionByHash())
	s.router.Methods("GET").Path("/{blockchain}/block/{number}").HandlerFunc(s.handlerBlockByNumber())
	s.router.Methods("GET").Path("/{blockchain}/address/{hash}/balance").HandlerFunc(s.handlerAddressBalance())
	s.router.Methods("GET").Path("/{blockchain}/address/{hash}/transactions").HandlerFunc(s.handlerAddressTransactions())
	s.router.Methods("GET").Path("/blockchain}/address/{hash}/owner").HandlerFunc(s.handlerAddressOwner())
	s.router.Methods("POST").Path("/blockchain}/address/{hash}/owner").HandlerFunc(s.handlerPostAddressOwner())

	return s, nil
}

func (s *Server) Listen(addr string) error {
	srv := &http.Server{
		Handler:      s.router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go s.ethClient.Run()

	server := &http.Server{Addr: ":8080", Handler: srv.Handler}
	return server.ListenAndServe()
}
