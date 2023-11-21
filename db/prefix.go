package db

import "fmt"

func prefixStatus() []byte {
	return []byte(fmt.Sprintf("status"))
}

func prefixBlock(number uint64) []byte {
	return []byte(fmt.Sprintf("block:%d", number))
}

func prefixTxByHash(hash string) []byte {
	return []byte(fmt.Sprintf("tx_hash:%s", hash))
}

func prefixTxBySender(address, hash string) []byte {
	return []byte(fmt.Sprintf("tx_sender:%s:%s", address, hash))
}

func prefixTxByReceiver(address, hash string) []byte {
	return []byte(fmt.Sprintf("tx_receiver:%s:%s", address, hash))
}

func prefixTxByBlockNumber(number uint64, hash string) []byte {
	return []byte(fmt.Sprintf("tx_block:%d:%s", number, hash))
}

func prefixGetTxByBlockNumber(number uint64) []byte {
	return []byte(fmt.Sprintf("tx_block:%d:", number))
}

func prefixGetTxBySender(to string) []byte {
	return []byte(fmt.Sprintf("tx_sender:%s", to))
}
