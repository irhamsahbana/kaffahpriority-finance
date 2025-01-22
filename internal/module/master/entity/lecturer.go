package entity

import "codebase-app/pkg/types"

type GetLecturersReq struct {
	Q string `query:"q" validate:"omitempty,min=3"`
	types.MetaQuery
}

func (r *GetLecturersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Lecturer struct {
	Common
	Phone *string `json:"phone" db:"phone"`
}

type GetLecturersResp struct {
	Items []Lecturer `json:"items"`
	Meta  types.Meta `json:"meta"`
}
