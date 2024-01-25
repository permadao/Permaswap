package account

import (
	"strconv"
	"testing"

	ethAccout "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/everFinance/goether"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	accId := "0xf9891e1a2635cb8d8c25a6a2ec8e453bfb2e67c4"
	_, id, _ := IDCheck(accId)
	// 1. success new
	acc, err := New(accId)
	assert.NoError(t, err)
	assert.Equal(t, id, acc.ID)
	assert.Equal(t, AccountTypeEVM, acc.Type)
	assert.Equal(t, int64(0), acc.Nonce)

	// 2. incorrect account id
	failedAccId := "0x111110000"
	acc, err = New(failedAccId)
	assert.Error(t, ERR_INVALID_ID)
}

func TestAccount_Verify(t *testing.T) {
	address01 := "0x3D7e9DFbc58952FdACEe2a5C69367C8478474D82"
	privKey01 := "ad1dcf8f1c449e7af21a7b8341eba5f053055819fff9948f1251ea94a0184cae"
	signer, err := goether.NewSigner(privKey01)
	assert.NoError(t, err)
	assert.Equal(t, address01, signer.Address.String())

	// 1. correct
	acc01, err := New(address01)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), acc01.Nonce)

	signData := []byte("this is a fake tx sign data")
	signHash := ethAccout.TextHash(signData)
	signedBytes, err := signer.SignMsg(signData)
	assert.NoError(t, err)

	tx01 := Transaction{
		Nonce: strconv.Itoa(int(acc01.Nonce + 1)), // tx nonce must > account nonce
		Hash:  signHash,
		Sig:   hexutil.Encode(signedBytes),
	}

	nonce, err := acc01.Verify(tx01)
	assert.NoError(t, err)
	assert.Equal(t, acc01.Nonce+1, nonce)

	// 2. nonce invalid
	tx02 := Transaction{
		Nonce: "",
		Hash:  nil,
		Sig:   "",
	}
	nonce, err = (&Account{}).Verify(tx02)
	assert.Error(t, ERR_INVALID_NONCE, err)
	assert.Equal(t, int64(0), nonce)

	// 3. nonce too low
	acc03, err := New(address01)
	assert.NoError(t, err)
	acc03.UpdateNonce(9) // update nonce from 0 to 9
	tx := Transaction{
		Nonce: "1",
		Hash:  nil,
		Sig:   "",
	}
	_, err = acc03.Verify(tx)
	assert.Error(t, ERR_NONCE_TOO_LOW, err)

	// 4. signature err
	// 4.1 test address different
	// fix step 1 signer and account is different address
	diffAddress01 := "0xF9891E1A2635CB8D8C25A6A2ec8E453bFb2E67c4"
	fakeAcc, err := New(diffAddress01)
	assert.NoError(t, err)

	_, err = fakeAcc.Verify(tx01)
	assert.Error(t, ERR_INVALID_SIGNATURE, err)

	// 4.2 test hash and sig
	// fix step 1 signHash for test
	fakeSignHash := ethAccout.TextHash(append(signData, 'a')) // add char 'a' to change sign hash
	tx01.Hash = fakeSignHash
	_, err = acc01.Verify(tx01)
	assert.Error(t, ERR_INVALID_SIGNATURE, err)
}

func TestArVerify(t *testing.T) {
	accid := "5NPqYBdIsIpJzPeYixuz7BEH_W7BEk_mb8HxBD3OHXo"
	acc, err := New(accid)
	assert.NoError(t, err)

	sig := "Z0pR_9-jHHr4P509cQUPadIppdk-nCHM7YPA8VaLw-zUb9QjRFkwiuZ0hUbv0WtHOhKW2rWd-cfCrbmkUm8BsfKB4RRSsN58BOtZ89_DH4qrZmkKEs__ymqQJ6GRJX8Wpzc5sNHmSlKvVGUbO0VIMgwVJgLwBnTkRxnmsHzj2_l5D-ENjSQs-L3_BP2QKKtNt4kvKAICVlBeVBwetsqsZjwkBONWExrqK8wFOTSy8g20NtjVheh2bj7uK-8Oa9w3ETXmlKkxhMq7Qj0Nj1Y0pXU3fYog5RER2JM8VB3Azn5z94iI7yaKcIio23Ap1-Kew14ZBbrfBDWPjfUlBquvam-0WZ14dvw8DUIv5ITfiz-ZwN00gH2OoM-S8jCjXRtt7zzQVMU704_3OAja9yJT62XHEEasKYnx9N6npEnC6aIfSNfGv2Cx4Yy4OKefWnrzWjzvwTs3dtgGHTbc135CVNLPYuQ87hFg0mgasKvG1aWHw7xy2nhme6VzMea2N3aF1x94tlFZe7Mog-lQG4ZYk-LEm1xCiYAifP-y58aaW5dzWp6xtEyuNw2M2l45_eQMGmhPv5yKBbqV6smAqjZNSiXYBGoFpviAAcFF-q1_U1Dz7lsYsKfXVcfOViIrRWCjRWWx0LCyiV_2sviTdve4HhElb3NYvynljw6XYa_p8oY,odtNk97a4PARR0I8g3kQpzlFVmPg-udyjfl81fbTioyP2pEw5tP5A1-FVqR-QFFPskW-j7yAze5usYNWHEir7oVQ9d9bbkcZIDEPqwSTO1JoD1BKXeeBK0xsmiSgxeY7uuRXWdhXREhlmIMsV8ObakEeXdbbxbs89XaZHBuES7boASrRVDXRz_mhMu6u_58OdLeMwR3I1BCH6nphNGVOehA7GOOqEBvtesBset0bNaLCb0JpSg5ZW_0AGLP-XydzE3IPLLx4NQEEJY21y8fChxYM4jntI78l5hojp9NlmS69EXlj0PoMjsbaWaz9WtnZaMAbnaOGAHhv8Y_TNmBI0FHpqHaGPP906Mnrgdm3tl2L40EX-Q6-liNVkB56CmPxXzSesu-4x5LLYxQ-aX3W6Hj7RCDTacxqUJHzOrhJqXSx6Jx0t8CwyfReMgVv4p5t1C3OZ8yYbJ_H3LdkeriVniaC5jQdMyIJ6QBMzr1XdXIw9WuEG2kCIYtvOp2qDuu9o2SY-9W4Yv7VWRDfWO38xxR4ZO65MMAdZxeaZ4w8sK_owH46Wm0XoT3Al-LPypaeijWqlHEu4R8c2ersD3xkDvXC_lNtaQw_qyfI3UEH5fWupY4zhZeDGkvXQh32Fv4CxlZL58iUHv9SvR7p5LgBCC3AVUbn7Sqc4xPUCZMj-Tc"

	err = acc.VerifySig(Transaction{
		Nonce: "0",
		Hash:  common.FromHex("0x14e0c3d7d499afbb227f815bb5732122c872c23b252ee8c9bb4491d82172432d"),
		Sig:   sig,
	})
	assert.NoError(t, err)
}
