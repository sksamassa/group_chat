package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func GenerateRandomBytes(length int) []byte {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Println(err)
	}

	return randomBytes
}

func Bytes(n int) ([]byte, error) {
	randBytes := make([]byte, n)
	nRead, err := rand.Read(randBytes)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if nRead < n {
		return nil, fmt.Errorf("bytes: didn't read enough random bytes")
	}
	return randBytes, nil
}

func String(n int) (string, error) {
	bytes, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
