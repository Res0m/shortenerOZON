package generator

import (
	"crypto/rand"
	"math/big"
)

const (
	symbols       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	lengthSymbols = len(symbols)
)

func Generate(size int) (string, error) {
	result := make([]byte, size)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(lengthSymbols)))
		if err != nil {
			return "", err
		}
		result[i] = symbols[n.Int64()]
	}
	return string(result), nil
}
