package hvm

import (
	"encoding/json"

	"github.com/permadao/permaswap/halo/hvm/schema"
)

func DeepCopyTx(src, dst *schema.Transaction) error {
	if tmp, err := json.Marshal(&src); err != nil {
		return err
	} else {
		err = json.Unmarshal(tmp, dst)
		return err
	}
}

func InSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func RemoveFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
