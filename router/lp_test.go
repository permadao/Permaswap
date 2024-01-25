package router

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/everFinance/goether"
	"github.com/permadao/permaswap/router/schema"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func testGenLp() *websocket.Conn {
	url := "ws://localhost" + testPort + "/wslp"
	lp, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	return lp
}

func testAutoRegisterLp(signer *goether.Signer) (lp *websocket.Conn, res []byte) {
	lp = testGenLp()
	_, msg, _ := lp.ReadMessage()
	// decode salt
	saltMsg := schema.LpMsgSalt{}
	json.Unmarshal(msg, &saltMsg)

	// register lp
	salt := saltMsg.Salt
	sig, _ := signer.SignMsg([]byte(salt))
	regMsg := schema.LpMsgRegister{
		Address: signer.Address.Hex(),
		Sig:     common.Bytes2Hex(sig),
	}
	lp.WriteMessage(websocket.TextMessage, regMsg.Marshal())
	_, res, _ = lp.ReadMessage()

	return
}

func TestLpRegister(t *testing.T) {
	testRouter := testGenRouter()
	defer func() {
		testRouter.Close()
	}()
	testEthSigner, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	testLp := testGenLp()

	_, msg, err := testLp.ReadMessage()
	assert.NoError(t, err)
	// decode salt
	saltMsg := schema.LpMsgSalt{}
	err = json.Unmarshal(msg, &saltMsg)
	assert.NoError(t, err)
	assert.Equal(t, schema.LpMsgEventSalt, saltMsg.Event)

	// WsErr
	var wsErr WsErr

	// invalid msg
	testLp.WriteMessage(websocket.TextMessage, []byte("123"))
	// get & decode wserr
	_, msg, _ = testLp.ReadMessage()
	json.Unmarshal(msg, &wsErr)
	assert.Equal(t, WsErrInvalidMsg, wsErr)

	// no auth
	addMsg := schema.LpMsgAdd{}
	testLp.WriteMessage(websocket.TextMessage, addMsg.Marshal())
	// get & decode wserr
	_, msg, _ = testLp.ReadMessage()
	err = json.Unmarshal(msg, &wsErr)
	assert.NoError(t, err)
	assert.Equal(t, WsErrNoAuthorization, wsErr)

	// register lp
	salt := saltMsg.Salt
	sig, _ := testEthSigner.SignMsg([]byte(salt))
	regMsg := schema.LpMsgRegister{
		Address: testEthSigner.Address.Hex(),
		Sig:     common.Bytes2Hex(sig),
	}
	testLp.WriteMessage(websocket.TextMessage, regMsg.Marshal())
	_, msg, _ = testLp.ReadMessage()
	var lpMsgRes schema.LpMsgResponse
	err = json.Unmarshal(msg, &lpMsgRes)
	assert.NoError(t, err)
	assert.Equal(t, schema.LpMsgOk, lpMsgRes)

	sessionID := testRouter.lpAddrToID[testEthSigner.Address.Hex()]
	assert.Equal(t, testEthSigner.Address.Hex(), testRouter.lpIDtoAddr[sessionID])
	assert.True(t, testRouter.isLpByID(sessionID))
	assert.True(t, testRouter.isLpByAddr(testEthSigner.Address.Hex()))

	// duplicate register
	_, msg = testAutoRegisterLp(testEthSigner)
	json.Unmarshal(msg, &wsErr)
	assert.Equal(t, WsErrDuplicateRegistration, wsErr)
}

func TestLpUnregister(t *testing.T) {
	testRouter := testGenRouter()
	defer func() {
		testRouter.Close()
	}()
	testEthSigner, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")

	lp, _ := testAutoRegisterLp(testEthSigner)
	assert.Equal(t, 1, len(testRouter.lpIDtoAddr))
	assert.Equal(t, 1, len(testRouter.lpAddrToID))
	assert.Equal(t, 1, len(testRouter.lpSalt))
	err := lp.Close()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(testRouter.lpIDtoAddr))
	assert.Equal(t, 0, len(testRouter.lpAddrToID))
	assert.Equal(t, 0, len(testRouter.lpSalt))
}

func TestLpAddAndRemove(t *testing.T) {
	testRouter := testGenRouter()
	defer func() {
		testRouter.Close()
	}()
	testEthSigner, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	lp, _ := testAutoRegisterLp(testEthSigner)

	// add
	config := `{
		"tokenX": "ethereum-eth-0x0000000000000000000000000000000000000000",
		"tokenY": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"feeRatio": "0.003",
		"lowSqrtPrice": "0.000044721359549995793928183473374626",
		"currentSqrtPrice": "0.000054792195750516611345696978280080",
		"highSqrtPrice": "0.000063245553203367586639977870888654",
		"liquidity": "50000000000000000",
		"priceDirection": "both"
	}`
	addMsg := schema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg)
	lp.WriteMessage(websocket.TextMessage, addMsg.Marshal())
	// test in core
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, len(testRouter.core.Lps))
	assert.Equal(t, 1, len(testRouter.core.AddressToLpIDs))
	lpAddr := testEthSigner.Address.Hex()
	lpIDs := testRouter.core.AddressToLpIDs[lpAddr]
	assert.Equal(t, 1, len(lpIDs))
	coreLp := testRouter.core.Lps[lpIDs[0]]
	corePool := testRouter.core.Pools[coreLp.PoolID]
	assert.Equal(t, "0x7bd8bbec75143287a3ac339d7f3235f130dd8e779663cde432558852d6d33d80", corePool.ID())

	// remove
	removeMsg := schema.LpMsgRemove{}
	json.Unmarshal([]byte(config), &removeMsg)
	lp.WriteMessage(websocket.TextMessage, removeMsg.Marshal())
	// test in core
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(testRouter.core.Lps))
	assert.Equal(t, 0, len(testRouter.core.AddressToLpIDs[testEthSigner.Address.Hex()]))
}

func TestLpSign(t *testing.T) {
	//see TestUserSubmit
}
