package entity

import "codebase-app/pkg/types"

type GetLecturerProgramsReq struct {
	UserId string `validate:"required,ulid"`

	types.MetaQuery
	IsFinanceUpdated string `query:"is_finance_updated" validate:"omitempty,oneof=true false"`
}

func (r *GetLecturerProgramsReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetLecturerProgramsResp struct {
	Items []LecturerProgramItem `json:"items"`
	Meta  types.Meta            `json:"meta"`
}

type LecturerProgramItem struct {
	LecturerId   string             `json:"lecturer_id" db:"lecturer_id"`
	LecturerName string             `json:"lecturer_name" db:"lecturer_name"`
	Templates    []LecturerTemplate `json:"templates"`
}

type LecturerTemplate struct {
	LecturerId       string  `json:"-" db:"lecturer_id"`
	TemplateId       string  `json:"template_id" db:"template_id"`
	ProgramId        string  `json:"program_id" db:"program_id"`
	StudentId        string  `json:"student_id" db:"student_id"`
	MarketerId       string  `json:"marketer_id" db:"marketer_id"`
	ProgramName      string  `json:"program_name" db:"program_name"`
	StudentName      string  `json:"student_name" db:"student_name"`
	MarketerName     string  `json:"marketer_name" db:"marketer_name"`
	MonthlyFee       float64 `json:"monthly_fee" db:"monthly_fee"`
	IsFinanceUpdated bool    `json:"is_finance_updated" db:"is_finance_updated"`
}
