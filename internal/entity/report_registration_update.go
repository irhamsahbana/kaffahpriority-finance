package entity

import (
	"codebase-app/pkg/errmsg"
	"fmt"
)

type UpdateRegistrationReq struct {
	UserID string `validate:"required,ulid"`

	ID                    string       `params:"id" validate:"ulid"`
	ProgramId             string       `json:"program_id" validate:"ulid"`
	MarketerId            string       `json:"marketer_id" validate:"ulid"`
	LecturerId            *string      `json:"lecturer_id" validate:"omitempty,ulid"`
	StudentId             string       `json:"student_id" validate:"ulid"`
	ProgramFee            float64      `json:"program_fee" validate:"required"`
	AdministrationFee     *float64     `json:"administration_fee" validate:"omitempty,min=0"`
	FLFee                 *float64     `json:"foreign_learning_fee" validate:"omitempty,min=0"`
	NLFee                 *float64     `json:"night_learning_fee" validate:"omitempty,min=0"`
	MarketerCommissionFee float64      `json:"marketer_commission_fee" validate:"min=0"`
	OverpaymentFee        *float64     `json:"overpayment_fee" validate:"omitempty,min=0"`
	HRFee                 float64      `json:"hr_fee" validate:"min=0"`
	MarketerGiftsFee      float64      `json:"marketer_gifts_fee" validate:"min=0"`
	ClosingFeeForOffice   *float64     `json:"closing_fee_for_office" validate:"omitempty,min=0"`
	ClosingFeeForReward   *float64     `json:"closing_fee_for_reward" validate:"omitempty,min=0"`
	Students              []AddStudent `json:"additional_students" validate:"required,dive"`
	Days                  []int64      `json:"days" validate:"required,unique_in_slice,dive,min=1,max=7"`
	Notes                 *string      `json:"notes" validate:"omitempty,max=255"`
	NotesForCategory      *string      `json:"notes_for_category" validate:"omitempty,max=255"`
	IsITP                 bool         `json:"is_itp"`
	Category              string       `json:"category"`
	IsUpdateTemplate bool   `json:"is_update_template"`
	PaidAt           string `json:"paid_at" validate:"required,datetime=2006-01-02"`
	PaidAtTime       string `json:"paid_at_time" validate:"required,datetime=15:04:05"`
	AllocatedAt      string `json:"allocated_at" validate:"required,datetime=2006-01-02"`
	AllocatedAtTime  string `json:"allocated_at_time" validate:"required,datetime=15:04:05"`
}

func (r *UpdateRegistrationReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	for i, s := range r.Students {
		if s.StudentID != nil && s.Name != nil {
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

type UpdateRegistrationResp struct {
	ID string `json:"id"`
}
