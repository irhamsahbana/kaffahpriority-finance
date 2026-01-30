package entity

import (
	"codebase-app/pkg/types"

	"github.com/lib/pq"
)

type GetTemplatesReq struct {
	Q      string `query:"q" validate:"omitempty,min=3"`
	UserID string `validate:"required,ulid"`

	types.MetaQuery

	MarketerId        string `query:"marketer_id" validate:"omitempty,ulid"`
	StudentManagerId  string `query:"student_manager_id" validate:"omitempty,ulid"`
	LecturerId        string `query:"lecturer_id" validate:"omitempty,ulid"`
	AcademicManagerId string `query:"academic_manager_id" validate:"omitempty,ulid"`

	StudentId                  string `query:"student_id" validate:"omitempty,ulid"`
	ProgramId                  string `query:"program_id" validate:"omitempty,ulid"`
	IsMandatoryFieldsCompleted *bool  `query:"is_mandatory_fields_completed" validate:"omitempty"`
}

func (r *GetTemplatesReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type GetTemplatesResp struct {
	Items []TemplateItem `json:"items"`
	Meta  types.Meta     `json:"meta"`
}

type TemplateItem struct {
	ID                      string        `json:"id" db:"id"`
	UserID                  string        `json:"user_id" db:"user_id"`
	ProgramID               string        `json:"program_id" db:"program_id"`
	MarketerID              *string       `json:"marketer_id" db:"marketer_id"`
	StudentManagerID        *string       `json:"student_manager_id" db:"student_manager_id"`
	LecturerID              *string       `json:"lecturer_id" db:"lecturer_id"`
	AcademicManagerID       *string       `json:"academic_manager_id" db:"academic_manager_id"`
	StudentID               string        `json:"student_id" db:"student_id"`
	StudentIdentifier       string        `json:"student_identifier" db:"student_identifier"`
	ProgramName             string        `json:"program_name" db:"program_name"`
	LecturerName            *string       `json:"lecturer_name" db:"lecturer_name"`
	AcademicManagerName     *string       `json:"academic_manager_name" db:"academic_manager_name"`
	MarketerName            *string       `json:"marketer_name" db:"marketer_name"`
	StudentManagerName      *string       `json:"student_manager_name" db:"student_manager_name"`
	StudentName             string        `json:"student_name" db:"student_name"`
	MonthlyFee              float64       `json:"monthly_fee" db:"monthly_fee"`
	Students                []AddStudent  `json:"additional_students"`
	Days                    pq.Int64Array `json:"days" db:"days"`
	ProgramFee              *float64      `json:"program_fee" db:"program_fee"`
	AdministrationFee       *float64      `json:"administration_fee" db:"administration_fee"`
	FLFee                   *float64      `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NLFee                   *float64      `json:"night_learning_fee" db:"night_learning_fee"`
	IsITP                   bool          `json:"is_itp" db:"is_itp"`
	MarketerCommissionFee   *float64      `json:"marketer_commission_fee" db:"marketer_commission_fee"`
	OverpaymentFee          *float64      `json:"overpayment_fee" db:"overpayment_fee"`
	HRFee                   float64       `json:"hr_fee" db:"hr_fee"`
	MarketerGiftsFee        float64       `json:"marketer_gifts_fee" db:"marketer_gifts_fee"`
	ClosingFeeForOffice     *float64      `json:"closing_fee_for_office" db:"closing_fee_for_office"`
	ClosingFeeForReward     *float64      `json:"closing_fee_for_reward" db:"closing_fee_for_reward"`
	Notes                   *string       `json:"notes" db:"notes"`
	IsFinanceUpdateRequired bool          `json:"is_finance_update_required" db:"is_finance_update_required"`
	CreatedAt               string        `json:"created_at" db:"created_at"`
	UpdatedAt               string        `json:"updated_at" db:"updated_at"`
	DeletedAt               *string       `json:"deleted_at" db:"deleted_at"`
	HasRegistrationPaid     bool          `json:"has_registration_paid" db:"has_registration_paid"`
}
