package entity

import "codebase-app/pkg/types"

type GetMarketersReq struct {
	UserId string `validate:"ulid"`

	Q string `query:"q" validate:"omitempty,min=3"`
	types.MetaQuery
}

func (r *GetMarketersReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Marketer struct {
	Common
	StudentManagerId   string  `json:"student_manager_id" db:"student_manager_id"`
	StudentManagerName string  `json:"student_manager_name" db:"student_manager_name"`
	Email              *string `json:"email" db:"email"`
	Phone              *string `json:"phone" db:"phone"`
}

type GetMarketersResp struct {
	Items []Marketer `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetMarketerReq struct {
	UserId string `validate:"ulid"`
	Marketer
}

type GetMarketerResp struct {
	Marketer
}

type CreateMarketerReq struct {
	UserId string `validate:"ulid"`

	StudentManagerId string  `json:"student_manager_id" validate:"required,ulid"`
	Name             string  `json:"name" validate:"required,min=3,max=255"`
	Email            *string `json:"email" validate:"omitempty,email"`
	Phone            *string `json:"phone" validate:"omitempty,min=9"`
}

type CreateMarketerResp struct {
	Id string `json:"id"`
}

type UpdateMarketerReq struct {
	UserId string `validate:"ulid"`

	Id               string  `json:"id" validate:"ulid"`
	StudentManagerId string  `json:"student_manager_id" validate:"required,ulid"`
	Name             string  `json:"name" validate:"required,min=3,max=255"`
	Email            *string `json:"email" validate:"omitempty,email"`
	Phone            *string `json:"phone" validate:"omitempty,min=9"`
}

type DeleteMarketerReq struct {
	UserId string `validate:"ulid"`

	Id string `json:"id" validate:"required,ulid"`
}
