package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) printChain(nodeID string) {
	// Создаем цепочку с имеющейся базой данных
	bc := NewBlockchain(nodeID)
	defer bc.db.Close()

	// Создаем итератор по цепочке
	bci := bc.Iterator()

	for {
		// Получаем новый блок
		block, hasNext := bci.Next()

		// Выводим информацию по блоку
		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)

		// Доказательство валидности блока
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		// Выводим транзакции блока
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		// Останавливаемся, если у нас больше нету хэша для итерирования
		if hasNext == false {
			break
		}
	}
}
