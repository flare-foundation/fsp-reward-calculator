package utils

import "math/big"

var (
	tmpA = new(big.Int)
	tmpB = new(big.Int)
)

// MinDec returns the numerically smaller of two strings.
func MinDec(a, b string) string {
	tmpA.SetString(a, 10)
	tmpB.SetString(b, 10)

	if tmpA.Cmp(tmpB) < 0 {
		return a
	}
	return b
}
