package rewards

import (
	"math/big"
)

const totalBips = 10000
const totalPpm = 1000000

var (
	bigTotalBips = big.NewInt(int64(totalBips))
	bigTotalPPM  = big.NewInt(int64(totalPpm))
	BigZero      = big.NewInt(0)
	// Used for temporary big.Int calculation results.
	bigTmp = new(big.Int)
)
