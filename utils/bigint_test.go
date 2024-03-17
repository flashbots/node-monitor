package utils_test

import (
	"math/big"
	"testing"

	"github.com/flashbots/node-monitor/utils"
	"gotest.tools/assert"
)

func TestBigint2string(t *testing.T) {
	zero := utils.Bigint2string(big.NewInt(0))
	assert.Equal(t, "00000000000000000000000000000000", zero)

	one := utils.Bigint2string(big.NewInt(1))
	assert.Equal(t, "00000000000000000000000000000001", one)
}
