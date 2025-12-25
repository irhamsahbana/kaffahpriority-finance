package entity

import "github.com/shopspring/decimal"

type RelatedRegistration struct {
	ID                  string           `json:"id" db:"id"`
	Category            string           `json:"category" db:"category"`
	LecturerID          *string          `json:"lecturer_id" db:"lecturer_id"`
	ProgramID           string           `json:"program_id" db:"program_id"`
	StudentID           string           `json:"student_id" db:"student_id"`
	MentorDetailFee     decimal.Decimal  `json:"mentor_detail_fee" db:"mentor_detail_fee"`
	MentorDetailFeeUsed *decimal.Decimal `json:"mentor_detail_fee_used" db:"mentor_detail_fee_used"`
	PaidAt              *string          `json:"paid_at" db:"paid_at"`
	AllocatedAt         *string          `json:"allocated_at" db:"allocated_at"`
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
