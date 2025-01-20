package entity

import (
	"codebase-app/pkg/errmsg"
	"fmt"
)

type CreateTemplateReq struct {
	UserId string ` json:"user_id" validate:"ulid"`

	ProgramId          string       `json:"program_id" validate:"ulid"`
	MarketerId         string       `json:"marketer_id" validate:"ulid"`
	LecturerId         string       `json:"lecturer_id" validate:"ulid"`
	StudentId          string       `json:"student_id" validate:"ulid"`
	AdditionalStudents []AddStudent `json:"additional_students" validate:"required,dive"`
	Days               []int        `json:"days" validate:"required,unique_in_slice,dive,min=1,max=7"`
}

func (req *CreateTemplateReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	for i, s := range req.AdditionalStudents {
		if s.StudentId != nil && s.Name != nil {
			err.Add(fmt.Sprintf("additional_students[%d].student_id", i), "student_id dan name tidak boleh diisi bersamaan")
			err.Add(fmt.Sprintf("additional_students[%d].name", i), "student_id dan name tidak boleh diisi bersamaan")
		}
	}

	if err.HasErrors() {
		return err
	} else {
		return nil
	}
}

type CreateTemplateResp struct {
	Id string `json:"id"`
}
