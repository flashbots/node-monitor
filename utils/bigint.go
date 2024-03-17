package utils

import (
	"math/big"
	"strings"
)

const (
	length = 32
)

func Bigint2string(bi *big.Int) string {
	str := bi.String()
	str = strings.Repeat("0", length-len(str)) + str
	return str
}
