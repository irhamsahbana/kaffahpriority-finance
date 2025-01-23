package entity

import (
	"codebase-app/pkg/types"

	"github.com/lib/pq"
)

type GetProgramsReq struct {
	types.MetaQuery
	Q string `query:"q" validate:"omitempty,min=2"`
}

func (r *GetProgramsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Program struct {
	Common
	Detail        *string       `json:"detail" db:"detail"`
	Price         float64       `json:"price" db:"price"`
	CommissionFee float64       `json:"commission_fee" db:"commission_fee"`
	LecturerFee   float64       `json:"lecturer_fee" db:"lecturer_fee"`
	Profit        float64       `json:"profit" db:"profit"`
	Days          pq.Int64Array `json:"days" db:"days"`
}

type GetProgramsResp struct {
	Items []Program  `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetProgramReq struct {
	Id string `params:"id" validate:"ulid"`
}

type GetProgramResp struct {
	Program
}

type CreateProgramReq struct {
	Name          string  `json:"name" validate:"required,min=3"`
	Detail        *string `json:"detail" validate:"omitempty,min=3"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	CommissionFee float64 `json:"commission_fee" validate:"required,gt=0"`
	LecturerFee   float64 `json:"lecturer_fee" validate:"required,gt=0"`
	Days          []int64 `json:"days" validate:"required,min=1,dive,min=1,max=7"`
}

type CreateProgramResp struct {
	Id string `json:"id"`
}
