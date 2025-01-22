package entity

import (
	"codebase-app/pkg/types"

	"github.com/lib/pq"
)

type Common struct {
	Id   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

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

type GetStudentsReq struct {
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

type GetProgramsReq struct {
	types.MetaQuery
	Q string `query:"q" validate:"omitempty,min=2"`
}

func (r *GetProgramsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type Program struct {
	Common
	Price         float64       `json:"price" db:"price"`
	CommissionFee float64       `json:"commission_fee" db:"commission_fee"`
	LecturerFee   float64       `json:"lecturer_fee" db:"lecturer_fee"`
	Profit        float64       `json:"profit" db:"profit"`
	Days          pq.Int64Array `json:"days" db:"days"`
}

type GetProgramsResp struct {
	Items []Program  `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type GetProgramReq struct {
	Id string `params:"id" validate:"ulid"`
}

type GetProgramResp struct {
	Program
}

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
