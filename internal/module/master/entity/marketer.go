package entity

import "codebase-app/pkg/types"

type GetMarketersReq struct {
	Q string `query:"q" validate:"omitempty,min=3"`
	types.MetaQuery
}

func (r *GetMarketersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Marketer struct {
	Common
	StudentManagerId string  `json:"student_manager_id" db:"student_manager_id"`
	StudentManager   string  `json:"student_manager_name" db:"student_manager_name"`
	Phone            *string `json:"phone" db:"phone"`
}

type GetMarketersResp struct {
	Items []Marketer `json:"items"`
	Meta  types.Meta `json:"meta"`
}
