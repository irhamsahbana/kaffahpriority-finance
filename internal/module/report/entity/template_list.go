package entity

import (
	"codebase-app/pkg/types"

	"github.com/lib/pq"
)

type GetTemplatesReq struct {
	types.MetaQuery
	IsFinanceUpdateRequired string `query:"is_finance_update_required" validate:"omitempty,oneof=true false"`
}

func (r *GetTemplatesReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetTemplatesResp struct {
	Items []TemplateItem `json:"items"`
	Meta  types.Meta     `json:"meta"`
}

type TemplateItem struct {
	Id                    string        `json:"id" db:"id"`
	UserId                string        `json:"user_id" db:"user_id"`
	ProgramId             string        `json:"program_id" db:"program_id"`
	MarketerId            string        `json:"marketer_id" db:"marketer_id"`
	LecturerId            string        `json:"lecturer_id" db:"lecturer_id"`
	StudentId             string        `json:"student_id" db:"student_id"`
	ProgramName           string        `json:"program_name" db:"program_name"`
	LecturerName          string        `json:"lecturer_name" db:"lecturer_name"`
	MarketerName          string        `json:"marketer_name" db:"marketer_name"`
	StudentName           string        `json:"student_name" db:"student_name"`
	MonthlyFee            float64       `json:"monthly_fee" db:"monthly_fee"`
	Students              []AddStudent  `json:"additional_students"`
	Days                  pq.Int64Array `json:"days" db:"days"`
	ProgramFee            *float64      `json:"program_fee" db:"program_fee"`
	AdministrationFee     *float64      `json:"administration_fee" db:"administration_fee"`
	FLFee                 *float64      `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NLFee                 *float64      `json:"night_learning_fee" db:"night_learning_fee"`
	MarketerCommissionFee *float64      `json:"marketer_commission_fee" db:"marketer_commission_fee"`
	OverpaymentFee        *float64      `json:"overpayment_fee" db:"overpayment_fee"`
	HRFee                 float64       `json:"hr_fee" db:"hr_fee"`
	MarketerGiftsFee      float64       `json:"marketer_gifts_fee" db:"marketer_gifts_fee"`
	ClosingFeeForOffice   *float64      `json:"closing_fee_for_office" db:"closing_fee_for_office"`
	ClosingFeeForReward   *float64      `json:"closing_fee_for_reward" db:"closing_fee_for_reward"`
	Notes                 *string       `json:"notes" db:"notes"`
	CreatedAt             string        `json:"created_at" db:"created_at"`
	UpdatedAt             string        `json:"updated_at" db:"updated_at"`
	DeletedAt             *string       `json:"deleted_at" db:"deleted_at"`
}
