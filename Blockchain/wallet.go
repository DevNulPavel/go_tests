package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet хранит приватный и публичный ключ нашего кошелька
type Wallet struct {
	PrivateKey ecdsa.PrivateKey // Закрытый ключ
	PublicKey  []byte           // Публичный ключ
}

// NewWallet создает и возвращает новый кошелек
func NewWallet() *Wallet {
	// Создаем приватный и публичный ключи
	private, public := newKeyPair()
	// Создаем кошелек с ними
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress возвращает адрес кошелька
func (w Wallet) GetAddress() []byte {
	// Получаем хэш от публичного ключа
	pubKeyHash := HashPubKey(w.PublicKey)

	// Слепляем версию и получившийся хэш, чтобы начиналось все с нуля
	versionedPayload := append([]byte{version}, pubKeyHash...)
	// Вычисляем контрольную сумму от результата
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey хэширует публичный ключ
func HashPubKey(pubKey []byte) []byte {
	// Берем публичный ключ
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0

	// Временно сделано для проверки
	//return true
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
