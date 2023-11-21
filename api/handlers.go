package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jkeddari/jarvis/types"
)

var serverInfo = types.ServerInfo{
	Infos: types.BlockchainsInfo{
		{
			Name: types.Ethereum,
			Symbols: []types.Symbol{
				types.ETH,
				types.USDT,
			},
		},
	},
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	s.Logger.Error().Err(err)
	apiError := types.APIError{
		Error: err.Error(),
	}

	w.WriteHeader(http.StatusInternalServerError)
	data, _ := json.Marshal(apiError)
	w.Write(data)
}

func (s *Server) writeJSON(w http.ResponseWriter, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		s.writeError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handlerStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, serverInfo)
	}
}

func (s *Server) handlerBlockchainStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		blockchain := vars["blockchain"]

		var status *types.BlockchainStatus
		var err error

		switch types.Blockchain(blockchain) {
		case types.Ethereum:
			status, err = s.ethClient.GetStatus()
		}

		if err != nil {
			s.writeError(w, err)
			return
		}

		s.writeJSON(w, status)
	}
}

func (s *Server) handlerTransactionByHash() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hash := vars["hash"]

		tx, err := s.ethClient.GetTransactionByHash(hash)
		if err != nil {
			s.writeError(w, err)
			return
		}

		s.writeJSON(w, tx)
	}
}

func (s *Server) handlerTransactionsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO : implement feature
	}
}

func (s *Server) handlerBlockByNumber() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		number, err := strconv.ParseUint(vars["number"], 10, 64)
		if err != nil {
			s.writeError(w, err)
			return
		}
		block, err := s.ethClient.GetBlockByNumber(number)
		if err != nil {
			s.writeError(w, err)
			return
		}

		s.writeJSON(w, block)
	}
}

func (s *Server) handlerAddressBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		ethBalance, err := s.ethClient.GetBalance(vars["hash"])
		if err != nil {
			s.Logger.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.writeJSON(w, ethBalance)
	}
}

func (s *Server) handlerAddressTransactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}

func (s *Server) handlerAddressOwner() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}
