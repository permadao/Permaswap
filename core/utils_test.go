package core

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFee(t *testing.T) {
	feeRatio, _ := StringToDecimal("0.003")

	fee, err := getFee(big.NewInt(1000), feeRatio, true)
	assert.NoError(t, err)
	assert.Equal(t, fee.String(), "3")

	fee, err = getFee(big.NewInt(1000), feeRatio, false)
	assert.Equal(t, fee.String(), "3")

	fee, err = getFee(big.NewInt(1), feeRatio, true)
	assert.Equal(t, fee.String(), "1")

	fee, err = getFee(big.NewInt(1), feeRatio, false)
	assert.Equal(t, fee.String(), "0")

	fee, err = getAndCheckFee(big.NewInt(1), feeRatio)
	assert.EqualError(t, err, "err_invalid_amount")
}
