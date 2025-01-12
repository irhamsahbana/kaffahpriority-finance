package entity

import "codebase-app/pkg/types"

type Common struct {
	Id   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type GetMarketersReq struct {
	types.MetaQuery
}

func (r *GetMarketersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetMarketersResp struct {
	Items []Common   `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetLecturersReq struct {
	types.MetaQuery
}

func (r *GetLecturersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetLecturersResp struct {
	Items []Common   `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetStudentsReq struct {
	IsActive string `query:"is_active" validate:"omitempty,oneof=true false"`
	types.MetaQuery
}

func (r *GetStudentsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Student struct {
	Common
	IsActive      bool    `json:"is_active" db:"is_active"`
	RegisteredAt  *string `json:"registered_at" db:"registered_at"`
	LastPaymentAt *string `json:"last_payment_at"`
}

type GetStudentsResp struct {
	Items []Student  `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetProgramsReq struct {
	types.MetaQuery
}

func (r *GetProgramsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Program struct {
	Common
	Price         float64 `json:"price" db:"price"`
	NumOfMeetings int     `json:"number_of_meetings" db:"number_of_meetings"`
}

type GetProgramsResp struct {
	Items []Program  `json:"items"`
	Meta  types.Meta `json:"meta"`
}
