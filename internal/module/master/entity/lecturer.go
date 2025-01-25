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

type GetLecturerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}

type GetLecturerResp struct {
	Lecturer
}

type CreateLecturerReq struct {
	UserId string `validate:"ulid"`

	Name  string  `json:"name" validate:"required,min=3"`
	Phone *string `json:"phone" validate:"omitempty,min=9"`
}

type CreateLecturerResp struct {
	Id string `json:"id"`
}

type UpdateLecturerReq struct {
	UserId string `validate:"ulid"`

	Id    string  `params:"id" validate:"required"`
	Name  string  `json:"name" validate:"required,min=3"`
	Phone *string `json:"phone" validate:"omitempty,min=9"`
}

type DeleteLecturerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}
