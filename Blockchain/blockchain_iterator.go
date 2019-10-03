package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator используется чтобы итерироваться по блокам блокчейна
type BlockchainIterator struct {
	currentHash []byte   // Хэш текущий
	db          *bolt.DB // База данных
}

// Next возвращает следующий блок, начиная с tip, возвращает блок и наличие следующего блока
func (i *BlockchainIterator) Next() (*Block, bool) {
	if len(i.currentHash) == 0 {
		return nil, false
	}

	var block *Block

	// Получаем доступ к базе данных на чтение
	err := i.db.View(func(tx *bolt.Tx) error {
		// Получаем доступ к корзине блоков
		b := tx.Bucket([]byte(blocksBucket))
		// Получаем данные блока для текущего хэша
		encodedBlock := b.Get(i.currentHash)
		// Создаем непосредственно блок из данных
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// Для следующей итерации сохраняем хэш предыдущего блока, по сути - обратный обход очереди получается
	i.currentHash = block.PrevBlockHash

	return block, (len(i.currentHash) != 0)
}
