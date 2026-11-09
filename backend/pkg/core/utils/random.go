package utils

import (
	"crypto/rand"
	"math/big"
)

// RandomString returns n characters drawn uniformly from alphabet with crypto/rand.
func RandomString(n int, alphabet string) (string, error) {
	chars := []rune(alphabet)
	max := big.NewInt(int64(len(chars)))
	out := make([]rune, n)
	for i := range out {
		k, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = chars[k.Int64()]
	}
	return string(out), nil
}
