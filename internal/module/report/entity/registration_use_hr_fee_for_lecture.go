package entity

import (
	"codebase-app/pkg/errmsg"

	"github.com/shopspring/decimal"
)

type UseHRfeeForLecturerReq struct {
	UserID string `json:"user_id" validate:"ulid"`

	RegistrationID string           `params:"registration_id" validate:"ulid"`
	UsedAmount     *decimal.Decimal `json:"used_amount"`
	Notes          *string          `json:"notes" validate:"omitempty,max=255"`
}

func (r *UseHRfeeForLecturerReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	if r.UsedAmount != nil && r.UsedAmount.LessThanOrEqual(decimal.Zero) {
		err.Add("used_amount", "used amount must be greater than 0")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type RelatedRegistration struct {
	ID                  string           `json:"id" db:"id"`
	LecturerID          *string          `json:"lecturer_id" db:"lecturer_id"`
	ProgramID           string           `json:"program_id" db:"program_id"`
	StudentID           string           `json:"student_id" db:"student_id"`
	MentorDetailFee     decimal.Decimal  `json:"mentor_detail_fee" db:"mentor_detail_fee"`
	MentorDetailFeeUsed *decimal.Decimal `json:"mentor_detail_fee_used" db:"mentor_detail_fee_used"`
	PaidAt              string           `json:"paid_at" db:"paid_at"`
	AllocatedAt         string           `json:"allocated_at" db:"allocated_at"`
}

type GetRelatedRegistrationsReq struct {
	UserID string `json:"user_id" validate:"ulid"`

	RegistrationID string `params:"registration_id" validate:"ulid"`
}

type GetRelatedRegistrationsResp struct {
	Items             []RelatedRegistration `json:"items"`
	TotalFee          decimal.Decimal       `json:"total_fee"`
	TotalFeeUsed      decimal.Decimal       `json:"total_fee_used"`
	TotalFeeRemaining decimal.Decimal       `json:"total_fee_remaining"`
}
