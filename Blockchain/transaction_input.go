package main

import "bytes"

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte // Хранит идентификатор такой транзакции, прошлый выход
	Vout      int    // Индекс выхода данной транзакции
	Signature []byte // Это скрипт, который предоставляет данные, которые будут в дальнейшем использоваться в скрипте ScriptPubKey
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
