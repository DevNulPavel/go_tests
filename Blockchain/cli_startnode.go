package main

import (
	"fmt"
	"log"
)

// Запуск нода
func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}

	// Запускаем сервер, который прослушивает подключающиеся соединения
	StartServer(nodeID, minerAddress)
}
