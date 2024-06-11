package utils

import (
	"github.com/pkg/errors"
	"math"
	"math/big"
)

var (
	maxUint32 = big.NewInt(math.MaxUint32)
)

func ToUint32(b *big.Int) (uint32, error) {
	if b.Sign() < 0 || b.Cmp(maxUint32) > 0 {
		return 0, errors.Errorf("value out of range for uint32: %v", b)
	}
	return uint32(b.Uint64()), nil
}
