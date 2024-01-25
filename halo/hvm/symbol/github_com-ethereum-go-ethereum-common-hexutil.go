// Code generated by 'yaegi extract github.com/ethereum/go-ethereum/common/hexutil'. DO NOT EDIT.

package symbol

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
)

func init() {
	Symbols["github.com/ethereum/go-ethereum/common/hexutil/hexutil"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Decode":                       reflect.ValueOf(hexutil.Decode),
		"DecodeBig":                    reflect.ValueOf(hexutil.DecodeBig),
		"DecodeUint64":                 reflect.ValueOf(hexutil.DecodeUint64),
		"Encode":                       reflect.ValueOf(hexutil.Encode),
		"EncodeBig":                    reflect.ValueOf(hexutil.EncodeBig),
		"EncodeUint64":                 reflect.ValueOf(hexutil.EncodeUint64),
		"ErrBig256Range":               reflect.ValueOf(&hexutil.ErrBig256Range).Elem(),
		"ErrEmptyNumber":               reflect.ValueOf(&hexutil.ErrEmptyNumber).Elem(),
		"ErrEmptyString":               reflect.ValueOf(&hexutil.ErrEmptyString).Elem(),
		"ErrLeadingZero":               reflect.ValueOf(&hexutil.ErrLeadingZero).Elem(),
		"ErrMissingPrefix":             reflect.ValueOf(&hexutil.ErrMissingPrefix).Elem(),
		"ErrOddLength":                 reflect.ValueOf(&hexutil.ErrOddLength).Elem(),
		"ErrSyntax":                    reflect.ValueOf(&hexutil.ErrSyntax).Elem(),
		"ErrUint64Range":               reflect.ValueOf(&hexutil.ErrUint64Range).Elem(),
		"ErrUintRange":                 reflect.ValueOf(&hexutil.ErrUintRange).Elem(),
		"MustDecode":                   reflect.ValueOf(hexutil.MustDecode),
		"MustDecodeBig":                reflect.ValueOf(hexutil.MustDecodeBig),
		"MustDecodeUint64":             reflect.ValueOf(hexutil.MustDecodeUint64),
		"UnmarshalFixedJSON":           reflect.ValueOf(hexutil.UnmarshalFixedJSON),
		"UnmarshalFixedText":           reflect.ValueOf(hexutil.UnmarshalFixedText),
		"UnmarshalFixedUnprefixedText": reflect.ValueOf(hexutil.UnmarshalFixedUnprefixedText),

		// type definitions
		"Big":    reflect.ValueOf((*hexutil.Big)(nil)),
		"Bytes":  reflect.ValueOf((*hexutil.Bytes)(nil)),
		"U256":   reflect.ValueOf((*hexutil.U256)(nil)),
		"Uint":   reflect.ValueOf((*hexutil.Uint)(nil)),
		"Uint64": reflect.ValueOf((*hexutil.Uint64)(nil)),
	}
}
