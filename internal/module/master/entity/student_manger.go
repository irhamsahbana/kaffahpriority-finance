package entity

import "codebase-app/pkg/types"

type GetStudentManagersReq struct {
	UserID string `validate:"ulid"`
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

type CreateStudentManagerReq struct {
	UserID string `validate:"ulid"`

	Name string `json:"name" validate:"required,min=3,max=255"`
}

type CreateStudentManagerResp struct {
	ID string `json:"id"`
}

type UpdateStudentManagerReq struct {
	UserID string `validate:"ulid"`

	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type DeleteStudentManagerReq struct {
	UserID string `validate:"ulid"`

	ID string `json:"id" validate:"required"`
}

type GetStudentManagerReq struct {
	UserID string `validate:"ulid"`

	ID string `json:"id" validate:"required"`
}

type GetStudentManagerResp struct {
	StudentManager
}
