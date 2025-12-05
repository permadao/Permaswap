package schema

import (
	ffpSchema "github.com/permadao/ffp/schema"
	"math/big"
)

const (
	TimeoutPeriod          = 300000                 // 5 minutes
	VmFFPAgentModuleFormat = "hymx.ffp.agent.0.0.1" // moduleId: UfrPioVwmdujZhO1I_CYhOzcdpwAuToER2gc900tSCU
)
const (
	AgentRequestPrice = "Agent-Request-Price" // pool agent action
	AgentPriceNotice  = "Agent-Price-Notice"  // user agent action
)

const (
	PoolAgentNoteStatusExecuting = "Executing"
	PoolAgentNoteStatusSettled   = "Settled"
	PoolAgentNoteStatusRefund    = "Refund"

	LimitOrderStatusPending      = "Pending"      // 待处理：订单已创建，等待价格满足
	LimitOrderStatusFailedPrice  = "FailedPrice"  // 已失败: 价格不满足
	LimitOrderStatusFailedSettle = "FailedSettle" // 已失败: 价格不满足
	LimitOrderStatusExecuting    = "Executing"    // 执行中：正在执行结算
	LimitOrderStatusCompleted    = "Completed"    // 已完成：结算成功
	LimitOrderStatusCancelled    = "Cancelled"    // 已取消：用户主动取消
	LimitOrderStatusExpired      = "Expired"      // 已过期：订单过期
)

type AgentNote struct {
	ffpSchema.Note
	Side   string // issuer or holder
	Status string `json:"Status"`
}

// LimitOrder
type LimitOrder struct {
	OrderID          string
	User             string
	TokenIn          string
	TokenOut         string
	AmountIn         *big.Int
	MinAmountOut     *big.Int
	MaxAmountOut     *big.Int
	Slippage         *big.Int
	PoolAgent        string
	Status           string
	SettledAmountOut string
	NoteId           string
	CreatedAt        int64
	RetryCount       int // 0, 1, 2
}

type OrderReq struct {
	TokenIn      string
	TokenOut     string
	AmountIn     *big.Int
	MaxAmountOut *big.Int
	Slippage     *big.Int
	PoolAgent    string
}

type SwapInputReq struct {
	Address   string // swap sender
	TokenIn   string
	TokenOut  string
	AmountIn  *big.Int
	AmountOut *big.Int
}
