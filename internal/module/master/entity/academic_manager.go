package entity

import "codebase-app/pkg/types"

type GetAcademicManagersReq struct {
	UserID string `validate:"ulid"`
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
	UserID string `validate:"ulid"`

	Name string `json:"name" validate:"required,min=3,max=255"`
}

type CreateAcademicManagerResp struct {
	ID string `json:"id"`
}

type UpdateAcademicManagerReq struct {
	UserID string `validate:"ulid"`

	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type DeleteAcademicManagerReq struct {
	UserID string `validate:"ulid"`

	ID string `json:"id" validate:"required"`
}

type GetAcademicManagerReq struct {
	UserID string `validate:"ulid"`

	ID string `json:"id" validate:"required"`
}

type GetAcademicManagerResp struct {
	AcademicManager
}
