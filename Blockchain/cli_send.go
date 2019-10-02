package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	// Проверяем валиднось адресов отправителя и получателя
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	// Создаем новый блок для текущего ID нода
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}

	// При завершении работы закрываем базу
	defer bc.db.Close()

	// Создаем кошелек для текущего нода
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	// Инициируем транзакцию
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	// Если надо майнить - стартуем вычисления, иначе откладываем
	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}
