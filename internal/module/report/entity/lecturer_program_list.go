package entity

import "codebase-app/pkg/types"

type GetLecturerProgramsReq struct {
	types.MetaQuery
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
	TemplateId   string  `json:"template_id" db:"template_id"`
	ProgramId    string  `json:"program_id" db:"program_id"`
	StudentId    string  `json:"student_id" db:"student_id"`
	MarketerId   string  `json:"marketer_id" db:"marketer_id"`
	ProgramName  string  `json:"program_name" db:"program_name"`
	StudentName  string  `json:"student_name" db:"student_name"`
	MarketerName string  `json:"marketer_name" db:"marketer_name"`
	MonthlyFee   float64 `json:"monthly_fee" db:"monthly_fee"`
}
