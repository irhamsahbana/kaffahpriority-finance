package entity

import "codebase-app/pkg/types"

type GetStudentManagersReq struct {
	UserId string `validate:"ulid"`
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
	UserId string `validate:"ulid"`

	Name string `json:"name" validate:"required,min=3,max=255"`
}

type CreateStudentManagerResp struct {
	Id string `json:"id"`
}

type UpdateStudentManagerReq struct {
	UserId string `validate:"ulid"`

	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type DeleteStudentManagerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}

type GetStudentManagerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}

type GetStudentManagerResp struct {
	StudentManager
}
