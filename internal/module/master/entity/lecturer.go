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
	AcademicManagerId   string  `json:"academic_manager_id" db:"academic_manager_id"`
	AcademicManagerName string  `json:"academic_manager_name" db:"academic_manager_name"`
	Phone               *string `json:"phone" db:"phone"`
	RegisteredAt        *string `json:"registered_at" db:"registered_at"`
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

	AcademicManagerId string  `json:"academic_manager_id" validate:"required,ulid"`
	Name              string  `json:"name" validate:"required,min=3"`
	Phone             *string `json:"phone" validate:"omitempty,min=9"`
	RegisteredAt      *string `json:"registered_at" validate:"omitempty,datetime=2006-01-02"`
}

type CreateLecturerResp struct {
	Id string `json:"id"`
}

type UpdateLecturerReq struct {
	UserId string `validate:"ulid"`

	Id                string  `params:"id" validate:"required"`
	AcademicManagerId string  `json:"academic_manager_id" validate:"required,ulid,exist=academic_managers.id"`
	Name              string  `json:"name" validate:"required,min=3"`
	Phone             *string `json:"phone" validate:"omitempty,min=9"`
	RegisteredAt      *string `json:"registered_at" validate:"omitempty,datetime=2006-01-02"`
}

type DeleteLecturerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required"`
}
