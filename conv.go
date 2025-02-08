package stylus

import (
	"math/big"
	"fmt"
)

func BUint256ToBig(x []byte) (*big.Int, error) {
	b := new(big.Int).SetBytes(x)
	if b.Cmp(new(big.Int)) > 0 {
		return b, nil
	} else {
		return nil, fmt.Errorf("calldata")
	}
}
