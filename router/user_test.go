package router

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/everFinance/goether"
	"github.com/permadao/permaswap/router/schema"
	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func testGenUser() *websocket.Conn {
	url := "ws://localhost" + testPort + "/wsuser"
	user, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	return user
}

func TestUserUnregisterd(t *testing.T) {
	r := testGenRouter()
	defer func() {
		r.Close()
	}()

	lpSigner, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	lp, _ := testAutoRegisterLp(lpSigner)

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

	u := testGenUser()
	qry := schema.UserMsgQuery{
		Address:  "0x61EbF673c200646236B2c53465bcA0699455d5FA",
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "3000000000",
	}
	u.WriteMessage(websocket.TextMessage, qry.Marshal())
	time.Sleep(100 * time.Millisecond)
	for _, uqt := range r.userQueryTag {
		assert.Equal(t, 1, len(uqt))
	}
	assert.Equal(t, 2, len(r.userQueryTag))

	// user close & clean userQueryTag
	u.Close()
	time.Sleep(100 * time.Millisecond)
	for _, uqt := range r.userQueryTag {
		assert.Equal(t, 0, len(uqt))
	}
}

func TestUserQuery(t *testing.T) {
	// router & add Lp
	r := testGenRouter()
	defer func() {
		r.Close()
	}()

	lpSigner, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	lp, _ := testAutoRegisterLp(lpSigner)

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

	u := testGenUser()
	userSigner, _ := goether.NewSigner("a612afcbce266637dc044b934dbef0e88ce91cceea8c8f9183f193b3a61d78e1")

	// query order
	qry := schema.UserMsgQuery{
		Address:  userSigner.Address.Hex(),
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "3000000000",
	}
	u.WriteMessage(websocket.TextMessage, qry.Marshal())
	_, msg, _ := u.ReadMessage()
	orderMsg := schema.UserMsgOrder{}
	json.Unmarshal([]byte(msg), &orderMsg)
	//t.Log("orderMsg:", orderMsg)
	assert.Equal(t, "3014.5059020188", orderMsg.Price)
}

func TestUserSubmit(t *testing.T) {
	testRouter := testGenRouter()
	defer func() {
		testRouter.Close()
	}()

	signer1, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	signer2, _ := goether.NewSigner("a612afcbce266637dc044b934dbef0e88ce91cceea8c8f9183f193b3a61d78e1")

	testLp, _ := testAutoRegisterLp(signer2)
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
	testLp.WriteMessage(websocket.TextMessage, addMsg.Marshal())

	testUser := testGenUser()
	testUser.WriteMessage(websocket.TextMessage, schema.UserMsgQuery{
		Address:  "0xa06b79E655Db7D7C3B3E7B2ccEEb068c3259d0C9",
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "3000000000",
	}.Marshal())

	_, msg, _ := testUser.ReadMessage()
	order := schema.UserMsgOrder{}
	json.Unmarshal(msg, &order)
	bundle := order.Bundle

	// not have signature
	submitMsg := schema.UserMsgSubmit{
		Event:   schema.UserMsgEventSubmit,
		Address: signer1.Address.Hex(),
		Bundle:  everSchema.BundleWithSigs{Bundle: bundle},
		Paths:   order.Paths,
	}
	testUser.WriteMessage(websocket.TextMessage, submitMsg.Marshal())
	_, msg, _ = testUser.ReadMessage()
	assert.Equal(t, WsErrInvalidOrder.Error(), string(msg))

	signData := []byte(bundle.String())
	sig1, _ := signer1.SignMsg(signData)
	submitMsg = schema.UserMsgSubmit{
		Event:   schema.UserMsgEventSubmit,
		Address: signer1.Address.Hex(),
		Bundle: everSchema.BundleWithSigs{
			Bundle: bundle,
			Sigs: map[string]string{
				signer1.Address.String(): hexutil.Encode(sig1),
			},
		},
		Paths: order.Paths,
	}
	testUser.WriteMessage(websocket.TextMessage, submitMsg.Marshal())
	_, msg, _ = testLp.ReadMessage()
	orderMsg := schema.LpMsgOrder{}
	json.Unmarshal(msg, &orderMsg)

	// lp sign
	sig2, _ := signer2.SignMsg([]byte(orderMsg.Bundle.String()))
	signMsg := schema.LpMsgSign{
		Event:   schema.LpMsgEventSign,
		Address: signer2.Address.Hex(),
		Bundle: everSchema.BundleWithSigs{
			Bundle: orderMsg.Bundle,
			Sigs: map[string]string{
				signer2.Address.String(): hexutil.Encode(sig2),
			},
		},
	}
	testLp.WriteMessage(websocket.TextMessage, signMsg.Marshal())
	// notice
	// time.Sleep(1 * time.Millisecond)
	_, msg, _ = testUser.ReadMessage()
	orderStatus := schema.OrderMsgStatus{}
	json.Unmarshal(msg, &orderStatus)
	assert.Equal(t, schema.OrderStatusSuccess, orderStatus.Status)
	_, msg, _ = testLp.ReadMessage()
	json.Unmarshal(msg, &orderStatus)
	assert.Equal(t, schema.OrderStatusSuccess, orderStatus.Status)
	// user get new order
	_, msg, _ = testUser.ReadMessage()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(testRouter.orders))
}

func TestUserSubmitWhenLpDisconnect(t *testing.T) {
	testRouter := testGenRouter()
	defer func() {
		testRouter.Close()
	}()

	signer1, _ := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	signer2, _ := goether.NewSigner("a612afcbce266637dc044b934dbef0e88ce91cceea8c8f9183f193b3a61d78e1")

	testLp, _ := testAutoRegisterLp(signer2)
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
	testLp.WriteMessage(websocket.TextMessage, addMsg.Marshal())

	testUser := testGenUser()
	testUser.WriteMessage(websocket.TextMessage, schema.UserMsgQuery{
		Address:  "0xa06b79E655Db7D7C3B3E7B2ccEEb068c3259d0C9",
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "3000000000",
	}.Marshal())

	_, msg, _ := testUser.ReadMessage()
	order := schema.UserMsgOrder{}
	json.Unmarshal(msg, &order)

	testLp.Close()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, len(testRouter.core.Lps))

	bundle := order.Bundle
	signData := []byte(bundle.String())
	sig1, _ := signer1.SignMsg(signData)
	submitMsg := schema.UserMsgSubmit{
		Event:   schema.UserMsgEventSubmit,
		Address: signer1.Address.Hex(),
		Bundle: everSchema.BundleWithSigs{
			Bundle: bundle,
			Sigs: map[string]string{
				signer1.Address.String(): hexutil.Encode(sig1),
			},
		},
		Paths: order.Paths,
	}
	testUser.WriteMessage(websocket.TextMessage, submitMsg.Marshal())
	_, msg, _ = testUser.ReadMessage()
	assert.Equal(t, NewWsErr("err_no_lp").Error(), string(msg))
}
