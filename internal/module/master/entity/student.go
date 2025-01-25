package entity

import "codebase-app/pkg/types"

type GetStudentsReq struct {
	UserId string `validate:"required,ulid"`

	IsActive string `query:"is_active" validate:"omitempty,oneof=true false"`
	Q        string `query:"q" validate:"omitempty,min=3"`
	types.MetaQuery
}

func (r *GetStudentsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Student struct {
	Common
	Identifier    string  `json:"identifier" db:"identifier"`
	IsActive      bool    `json:"is_active" db:"is_active"`
	RegisteredAt  *string `json:"registered_at" db:"registered_at"`
	LastPaymentAt *string `json:"last_payment_at"`
}

type GetStudentsResp struct {
	Items []Student  `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetStudentReq struct {
	UserId string `validate:"required,ulid"`

	Id string `params:"id" validate:"required,ulid"`
}

type GetStudentResp struct {
	Student
}

type CreateStudentReq struct {
	UserId string `validate:"required,ulid"`

	Identifier   string  `json:"identifier" validate:"required"`
	Name         string  `json:"name" validate:"required,min=3"`
	RegisteredAt *string `json:"registered_at" validate:"omitempty,datetime=2006-01-02"`
}

type CreateStudentResp struct {
	Id string `json:"id"`
}

type UpdateStudentReq struct {
	UserId string `validate:"required,ulid"`

	Id           string  `params:"id" validate:"required,ulid"`
	Identifier   string  `json:"identifier" validate:"required"`
	Name         string  `json:"name" validate:"required,min=3"`
	RegisteredAt *string `json:"registered_at" validate:"omitempty,datetime=2006-01-02"`
	IsActive     bool    `json:"is_active"`
}

type DeleteStudentReq struct {
	UserId string `validate:"required,ulid"`

	Id string `params:"id" validate:"required,ulid"`
}
