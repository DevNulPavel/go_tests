package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// Block - структура, представляющая отдельный блок в цепочке
type Block struct {
	Timestamp     int64          // Время создания блока
	Transactions  []*Transaction // Полезная информация
	PrevBlockHash []byte         // Хэш предыдущего блока
	Hash          []byte         // Хэш текущего блока
	Nonce         int            // Номер успешной попытки?
	Height        int            // Вес
}

// NewBlock Функция, которая возращает новый блок
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	// Создаем непосредственно объект блока
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}

	// Создаем объект для вычисления доказательства работы
	pow := NewProofOfWork(block)
	// Выполняем подсчет доказательства работы и хэша
	nonce, hash := pow.Run()

	// Сохраняем их в блоке
	block.Hash = hash[:] // TODO: Зачем [:]??
	block.Nonce = nonce

	return block
}

// NewGenesisBlock С помощью этой функции можно создавать базовый блок, без предыдущих
func NewGenesisBlock(coinbase *Transaction) *Block {
	// Создание базового блока без предварительного хэша
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// HashTransactions Возвращает хэш от вычисленных транзакций данного блока
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// Serialize Выполняем сериализацию блока в набор байтов
func (b *Block) Serialize() []byte {
	// Создаем буффер
	var result bytes.Buffer
	// Создаем кодировщик с результатами в этом буффере
	encoder := gob.NewEncoder(&result)

	// Кодируем наш блок в массив байтов данных
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	// Создаем переменную, в которую будем засовывать блок
	var block Block

	// Создаем декодер на основе ридера данных
	decoder := gob.NewDecoder(bytes.NewReader(d))
	// Декодируем
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
