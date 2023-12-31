package db

import (
	"bytes"
	"encoding/gob"

	"github.com/dgraph-io/badger/v4"
	"github.com/jkeddari/jarvis/types"
)

// Data storage model :
//
// - Status storage :
//	    "status" => types.Status
//
// - Block Storage :
//  	"block:{number}" => types.Block
//
// - Transaction storage :
//   	"tx_hash:{hash}" = > types.Transaction
//   	"tx_sender:{address}:{tx_hash}" => {tx_hash}
//  	"tx_receiver:{address}:{tx_hash}" => {tx_hash}
//  	"tx_block:{number}:{tx_hash}" => {tx_hash}

func encodeData(data any) ([]byte, error) {
	var buff bytes.Buffer
	e := gob.NewEncoder(&buff)
	if err := e.Encode(data); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func decodeData(dest any, data []byte) error {
	d := gob.NewDecoder(bytes.NewReader(data))
	return d.Decode(dest)
}

type badgerDB struct {
	db *badger.DB
}

// NewBadgerDB returns a new badger database object.
func NewBadgerDB(path string) (DB, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &badgerDB{
		db: db,
	}, nil
}

func (b *badgerDB) DropAll() error {
	return b.db.DropAll()
}

func (b *badgerDB) SetStatus(status types.BlockchainStatus) error {
	data, err := encodeData(status)
	if err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(prefixStatus(), data)
	})
}

func (b *badgerDB) AddTransactions(txs ...types.Transaction) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, tx := range txs {
			data, err := encodeData(tx)
			if err != nil {
				return err
			}

			if err := txn.Set(prefixTxByHash(tx.Hash), data); err != nil {
				return err
			}
			if err := txn.Set(prefixTxBySender(tx.From, tx.Hash), []byte(tx.Hash)); err != nil {
				return err
			}

			if err := txn.Set(prefixTxByBlockNumber(tx.BlockNumber, tx.Hash), []byte(tx.Hash)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *badgerDB) UpdateNumber(number uint64) error {
	status, err := b.Status()
	if err != nil {
		return err
	}
	status.BlockCount++
	if number > status.EndNumber {
		status.EndNumber = number
	}
	return b.SetStatus(*status)
}

func (b *badgerDB) AddBlock(block types.Block) error {
	data, err := encodeData(block)
	if err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(prefixBlock(block.Number), data)
	})
}

func (b *badgerDB) Status() (status *types.BlockchainStatus, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		data, err := txn.Get(prefixStatus())
		if err != nil {
			return err
		}
		return data.Value(func(val []byte) error {
			return decodeData(&status, val)
		})
	})
	return
}

func (b *badgerDB) TransactionsForBlock(number uint64) (txs types.Transactions, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := prefixGetTxByBlockNumber(number)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			item.Value(func(hash []byte) error {
				tx, err := b.TransactionByHash(string(hash))
				if err != nil {
					return err
				}
				txs = append(txs, *tx)
				return nil
			})
		}

		return nil
	})
	return
}

func (b *badgerDB) BlockByNumber(number uint64) (block *types.Block, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefixBlock(number))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return decodeData(&block, val)
		})
	})

	return
}

func (b *badgerDB) TransactionByHash(hash string) (tx *types.Transaction, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefixTxByHash(hash))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return decodeData(&tx, val)
		})
	})

	return
}

func (b *badgerDB) TransactionsForAddress(address string) (txs types.Transactions, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := prefixGetTxBySender(address)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(v []byte) error {
				hash := string(v)
				tx, err := b.TransactionByHash(hash)
				if err == nil {
					txs = append(txs, *tx)
				}
				return err
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func (b *badgerDB) SetAddressOwner(addressOwner types.AddressOwner) error {
	return b.db.Update(func(txn *badger.Txn) error {
		data, err := encodeData(addressOwner)
		if err != nil {
			return err
		}
		return txn.Set(prefixAddressOwner(addressOwner.Address), data)
	})
}

func (b *badgerDB) GetAddressOwner(address string) (addressOwner *types.AddressOwner, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefixAddressOwner(address))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return decodeData(addressOwner, val)
		})
	})
	return nil, nil
}
