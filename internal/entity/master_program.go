package entity

import (
	"codebase-app/pkg/types"

	"github.com/lib/pq"
)

type GetProgramsReq struct {
	UserID string `validate:"ulid"`

	Q string `query:"q" validate:"omitempty,min=2"`
	types.MetaQuery
}

func (r *GetProgramsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Program struct {
	Common
	Detail            *string `json:"detail" db:"detail"`
	Price             float64 `json:"price" db:"price"`
	PricePerMeeting   float64 `json:"price_per_meeting" db:"price_per_meeting"`
	FullFee           float64 `json:"full_fee" db:"full_fee"`
	AcquisitionRights int64   `json:"acquisition_rights" db:"acquisition_rights"`
	CommissionFee     float64 `json:"commission_fee" db:"commission_fee"`
	LecturerFee       float64 `json:"lecturer_fee" db:"lecturer_fee"`
	Profit            float64 `json:"profit" db:"profit"`

	Days pq.Int64Array `json:"days" db:"days"`
}

type GetProgramsResp struct {
	Items []Program  `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetProgramReq struct {
	UserID string `validate:"ulid"`
	ID     string `params:"id" validate:"ulid"`
}

type GetProgramResp struct {
	Program
}

type CreateProgramReq struct {
	UserID string `validate:"ulid"`

	Name              string  `json:"name" validate:"required,min=3"`
	Detail            *string `json:"detail" validate:"omitempty,min=3"`
	Price             float64 `json:"price" validate:"required,gte=0"`
	PricePerMeeting   float64 `json:"price_per_meeting" validate:"required,gte=0"`
	FullFee           float64 `json:"full_fee" validate:"required,gte=0"`
	AcquisitionRights *int64  `json:"acquisition_rights" validate:"required,gte=0"`
	CommissionFee     float64 `json:"commission_fee" validate:"required,gte=0"`
	LecturerFee       float64 `json:"lecturer_fee" validate:"required,gte=0"`
	Days              []int64 `json:"days" validate:"required,min=1,dive,min=1,max=7"`
}

type CreateProgramResp struct {
	ID string `json:"id"`
}

type UpdateProgramReq struct {
	UserID string `validate:"ulid"`

	ID                string  `params:"id" validate:"required,ulid"`
	Name              string  `json:"name" validate:"required,min=3"`
	Detail            *string `json:"detail" validate:"omitempty,min=3"`
	Price             float64 `json:"price" validate:"required,gt=0"`
	PricePerMeeting   float64 `json:"price_per_meeting" validate:"required,gt=0"`
	FullFee           float64 `json:"full_fee" validate:"required,gt=0"`
	AcquisitionRights *int64  `json:"acquisition_rights" validate:"required,gte=0"`
	CommissionFee     float64 `json:"commission_fee" validate:"required,gte=0"`
	LecturerFee       float64 `json:"lecturer_fee" validate:"required,gte=0"`
	Days              []int64 `json:"days" validate:"required,min=1,dive,min=1,max=7"`
}

type UpdateProgramResp struct {
	ID string `json:"id"`
}

type DeleteProgramReq struct {
	UserID string `validate:"ulid"`
	ID     string `params:"id" validate:"required,ulid"`
}

type DeleteProgramResp struct {
	ID string `json:"id"`
}
