package schema

import (
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	everSchema "github.com/everVision/everpay-kits/schema"
)

type Oracle struct {
	EverTokens map[string]everSchema.TokenInfo
}

type Executor struct {
	LocalState     string `json:"localState"`
	LocalStateHash string `json:"localStateHash"`
	RunnedTimes    int64  `json:"runnedTimes"`
	// in: tx, state, oracle, localState, initData out: state, localState, localStateHash error
	Execute func(*Transaction, *StateForProposal, *Oracle, string, string) (*StateForProposal, string, string, error) `json:"-"`
}

type Proposal struct {
	Name                  string   `json:"name"`
	ID                    string   `json:"id"`
	Start                 int64    `json:"start"`
	End                   int64    `json:"end"`
	RunTimes              int64    `json:"runTimes"`
	Source                string   `json:"source"`
	InitData              string   `json:"initData"`
	OnlyAcceptedTxActions []string `json:"onlyAcceptedTxActions"`

	Executor *Executor `json:"executor"`
}

func (p *Proposal) String() string {
	onlyAcceptedTxActions := strings.Join(p.OnlyAcceptedTxActions, ",")
	return "name:" + p.Name + "\n" +
		"start:" + strconv.FormatInt(p.Start, 10) + "\n" +
		"end:" + strconv.FormatInt(p.End, 10) + "\n" +
		"runTimes:" + strconv.FormatInt(p.RunTimes, 10) + "\n" +
		"source:" + p.Source + "\n" +
		"initData:" + p.InitData + "\n" +
		"onlyAcceptedTxActions:" + onlyAcceptedTxActions + "\n"
}

func (p *Proposal) Hash() []byte {
	return accounts.TextHash([]byte(p.String()))
}

func (p *Proposal) HexHash() string {
	return hexutil.Encode(p.Hash())
}
