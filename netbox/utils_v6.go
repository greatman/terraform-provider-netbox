package netbox

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func toInt64ListV6(a types.Set) []int64 {
	intList := []int64{}
	intElements := make([]types.Int64, 0, len(a.Elements()))
	for _, number := range intElements {
		intList = append(intList, number.ValueInt64())
	}
	return intList
}
