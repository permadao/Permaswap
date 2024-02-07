//go:generate yaegi extract github.com/permadao/permaswap/halo/schema
//go:generate yaegi extract github.com/permadao/permaswap/halo/hvm/schema
//go:generate yaegi extract github.com/permadao/permaswap/halo/token/schema
//go:generate yaegi extract	"github.com/permadao/permaswap/logger"
//go:generate yaegi extract "github.com/ethereum/go-ethereum/accounts"
//go:generate yaegi extract	"github.com/ethereum/go-ethereum/common/hexutil"
package symbol

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}
