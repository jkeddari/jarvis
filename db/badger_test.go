package db

import (
	"testing"

	"github.com/jkeddari/jarvis/types"
	"github.com/stretchr/testify/assert"
)

var txs = types.Transactions{
	{
		BlockNumber: 42,
		Hash:        "tx1",
		Timestamp:   0,
		Fee:         0,
		From:        "foo",
		To:          "bar",
		Amount:      40,
	},
	{
		BlockNumber: 42,
		Hash:        "tx2",
		Timestamp:   0,
		Fee:         0,
		From:        "foo",
		To:          "bar",
		Amount:      42,
	},
	{
		BlockNumber: 42,
		Hash:        "tx3",
		Timestamp:   0,
		Fee:         0,
		From:        "bar",
		To:          "foo",
		Amount:      45,
	},
}

var testBlock = &types.Block{
	Number:             42,
	Hash:               "djkskljsksjfkjdsmkfjsqmfdjqsmlfjqsmljfmklqsjdfklj",
	Timestamp:          12874,
	TransactionsNumber: len(txsBlock),
}

var txsBlock = types.Transactions{
	{
		BlockNumber: 42,
		Hash:        "mlkdmlkqmlkd",
		Timestamp:   0,
		Fee:         0,
		From:        "toto",
		To:          "bar",
		Amount:      100,
	},
}

func TestDB(t *testing.T) {
	db, err := NewBadgerDB("/tmp/jarvis_db")
	assert.Nil(t, err)
	assert.Nil(t, db.DropAll())
	t.Run("status", func(t *testing.T) {
		status := &types.BlockchainStatus{
			StartNumber: 100,
			EndNumber:   0,
		}

		err = db.SetStatus(*status)
		assert.Nil(t, err)

		s, err := db.Status()
		assert.Equal(t, status, s)

		err = db.UpdateNumber(101)
		assert.Nil(t, err)

		s, err = db.Status()
		assert.Equal(t, uint64(101), s.EndNumber)
	})
	t.Run("block", func(t *testing.T) {
		err = db.AddBlock(*testBlock)
		assert.Nil(t, err)

		b, err := db.BlockByNumber(42)
		assert.Nil(t, err)
		assert.Equal(t, testBlock, b)
	})

	t.Run("transactions", func(t *testing.T) {
		for _, tx := range txs {
			err = db.AddTransactions(tx)
			assert.Nil(t, err)
		}

		expectedTx1, err := db.TransactionByHash("tx1")
		assert.Nil(t, err)
		assert.Equal(t, txs[0], *expectedTx1)

		fromFoo, err := db.TransactionsForAddress("foo")
		assert.Nil(t, err)

		assert.Equal(t, txs[:2], fromFoo)
	})
}
