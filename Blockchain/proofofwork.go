package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 16

// ProofOfWork представляет собой систему расчета доказательства работы
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork создает и возвращает новое доказательство работ
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// Выполняем смещение 1цы на 256-16=240 бит
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	// Подготовка данных для расчета хэша, здесь мы сложим все данные блока
	// такие как:
	// хэш предыдущего блока
	// текущий хэш транзакций
	// время, количество бит и nonce
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run выполняет непосредственно расчет
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")
	// Проверяем, что номер попытки меньше максимального
	for nonce < maxNonce {
		// Создаем данные для расчета
		data := pow.prepareData(nonce)

		// Считаем от них хэш
		hash = sha256.Sum256(data)
		//fmt.Printf("\r%x", hash) // Вывод пока что отключим
		hashInt.SetBytes(hash[:])

		// Сравниваем, если хэш меньше целевого числа - ура, нашли - иначе пробуем с числом на 1цу больше
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate проверяет доказательство работы
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	// Берем совокупные даннные и считаем от них хэш
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	// Сравниваем, значение должно быть меньше целевого
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
