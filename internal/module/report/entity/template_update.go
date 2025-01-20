package entity

import (
	"codebase-app/pkg/errmsg"
	"fmt"
)

type UpdateTemplateGeneralReq struct {
	Id     string `params:"id" validate:"ulid"`
	UserId string ` json:"user_id" validate:"ulid"`

	ProgramId          string       `json:"program_id" validate:"ulid"`
	MarketerId         string       `json:"marketer_id" validate:"ulid"`
	LecturerId         string       `json:"lecturer_id" validate:"ulid"`
	StudentId          string       `json:"student_id" validate:"ulid"`
	AdditionalStudents []AddStudent `json:"additional_students" validate:"required,dive"`
	Days               []int        `json:"days" validate:"required,unique_in_slice,dive,min=1,max=7"`
}

func (req *UpdateTemplateGeneralReq) Validate() error {
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

type UpdateTemplateFinanceReq struct {
	Id string `params:"id" validate:"ulid"`

	ProgramFee            *float64 `json:"program_fee" validate:"required,min=0"`
	AdministrationFee     float64  `json:"administration_fee" validate:"min=0"`
	FLFee                 *float64 `json:"foreign_learning_fee" validate:"omitempty,min=0"`
	NLFee                 *float64 `json:"night_learning_fee" validate:"omitempty,min=0"`
	MarketerCommissionFee float64  `json:"marketer_commission_fee" validate:"min=0"`
	OverpaymentFee        *float64 `json:"overpayment_fee" validate:"omitempty,min=0"`
	HRFee                 float64  `json:"hr_fee" validate:"min=0"`
	MarketerGiftsFee      float64  `json:"marketer_gifts_fee" validate:"min=0"`
	ClosingFeeForOffice   *float64 `json:"closing_fee_for_office" validate:"omitempty,min=0"`
	ClosingFeeForReward   *float64 `json:"closing_fee_for_reward" validate:"omitempty,min=0"`
}

type UpdateTemplateResp struct {
	Id string `json:"id"`
}
