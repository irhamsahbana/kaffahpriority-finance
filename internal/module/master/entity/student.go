package entity

import "codebase-app/pkg/types"

type GetStudentsReq struct {
	IsActive string `query:"is_active" validate:"omitempty,oneof=true false"`
	Q        string `query:"q" validate:"omitempty,min=3"`
	types.MetaQuery
}

func (r *GetStudentsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Student struct {
	Common
	Identifier    string  `json:"identifier" db:"identifier"`
	IsActive      bool    `json:"is_active" db:"is_active"`
	RegisteredAt  *string `json:"registered_at" db:"registered_at"`
	LastPaymentAt *string `json:"last_payment_at"`
}

type GetStudentsResp struct {
	Items []Student  `json:"items"`
	Meta  types.Meta `json:"meta"`
}
