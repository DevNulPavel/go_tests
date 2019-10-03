package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address, nodeID string) {
	// Проверяем адрес на валидность
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	// Создаем новую цепучку с существующей базой для текущего нода
	bc := NewBlockchain(nodeID)
	// Создаем непотраченых выходов для цепочки
	UTXOSet := UTXOSet{bc}

	// Не забываем закрыть базу
	defer bc.db.Close()

	// Текущий баланс
	balance := 0
	// В качестве публичного ключа будем использовать кодированый от адрес
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] // Сокращаем его
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)          // Выполняем поиск непотраченых выходов

	// Суммируем баланс
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
