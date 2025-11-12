package contract

import (
	huma "github.com/danielgtaylor/huma/v2"
)

type NonNullIntArrayType []int

func (NonNullIntArrayType) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "array",
		Items: &huma.Schema{
			Type:   "integer",
			Format: "int64",
		},
	}
}
