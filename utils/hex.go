package utils

import "math/big"

var (
	tmpA = new(big.Int)
	tmpB = new(big.Int)
)

// MinHex returns the numerically smaller of two hex strings.
func MinHex(a, b string) string {
	tmpA.SetString(a, 16)
	tmpB.SetString(b, 16)

	if tmpA.Cmp(tmpB) < 0 {
		return a
	}
	return b
}
