package entity

import "codebase-app/pkg/types"

type GetStudentManagersReq struct {
	types.MetaQuery
}

func (r *GetStudentManagersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type StudentManager struct {
	Common
}

type GetStudentManagersResp struct {
	Items []StudentManager `json:"items"`
	Meta  types.Meta       `json:"meta"`
}
