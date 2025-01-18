package entity

import "codebase-app/pkg/types"

type GetTemplatesReq struct {
	types.MetaQuery
}

func (r *GetTemplatesReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetTemplatesResp struct {
	Items []TemplateItem `json:"items"`
	Meta  types.Meta     `json:"meta"`
}

type TemplateItem struct {
	Id           string        `json:"id" db:"id"`
	MarketerId   string        `json:"marketer_id" db:"marketer_id"`
	LecturerId   string        `json:"lecturer_id" db:"lecturer_id"`
	StudentId    string        `json:"student_id" db:"student_id"`
	LecturerName string        `json:"lecturer_name" db:"lecturer_name"`
	MarketerName string        `json:"marketer_name" db:"marketer_name"`
	StudentName  string        `json:"student_name" db:"student_name"`
	Students     []StudentItem `json:"additional_students"`
}

type StudentItem struct {
	Id   *string `json:"id" db:"id"`
	Name string  `json:"name" db:"name"`
}
