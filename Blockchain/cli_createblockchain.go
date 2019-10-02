package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	// Для первого создания базы нам не нужна валидация
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockchain(address, nodeID)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
