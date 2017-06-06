package helpers

import "crypto/rand"
import "fmt"

func GenerateStringId() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return fmt.Sprintf("%x", bytes)
}