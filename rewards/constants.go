package rewards

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const totalBips = 10000
const totalPpm = 1000000

var (
	burnAddress  = common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	bigTotalBips = big.NewInt(int64(totalBips))
	bigTotalPPM  = big.NewInt(int64(totalPpm))
	bigZero      = big.NewInt(0)
	// Used for temporary big.Int calculation results
	bigTmp = new(big.Int)
)
