package entity

import "codebase-app/pkg/types"

type GetAcademicManagersReq struct {
	UserId string `validate:"ulid"`
	Q      string `query:"q"`
	types.MetaQuery
}

func (r *GetAcademicManagersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type AcademicManager struct {
	Common
}

type GetAcademicManagersResp struct {
	Items []AcademicManager `json:"items"`
	Meta  types.Meta        `json:"meta"`
}

type CreateAcademicManagerReq struct {
	UserId string `validate:"ulid"`

	Name string `json:"name" validate:"required,min=3,max=255"`
}

type CreateAcademicManagerResp struct {
	Id string `json:"id"`
}

type UpdateAcademicManagerReq struct {
	UserId string `validate:"ulid"`

	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type DeleteAcademicManagerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}

type GetAcademicManagerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}

type GetAcademicManagerResp struct {
	AcademicManager
}
