package entity

import (
	"codebase-app/pkg/types"

	"github.com/shopspring/decimal"
)

type CreatePayrollRunReq struct {
	Period string `json:"period"` // e.g. "2025-01"
	UserID string `json:"user_id"`
}

type CreatePayrollRunResp struct {
	ID string `json:"id"`
}

type GetPayrollRunsReq struct {
	Q string `query:"q"`
	types.MetaQuery
}

type GetPayrollRunsResp struct {
	Items []PayrollRun `json:"items"`
	Meta  types.Meta   `json:"meta"`
}

type GetPayrollRunDetailReq struct {
	ID string `params:"id" validate:"required,ulid"`
}

type GetPayrollRunDetailResp struct {
	PayrollRun
	Items []PayrollItem `json:"items"`
}

type GetPayrollItemsReq struct {
	Period     string `query:"period"`
	Q          string `query:"q"`
	MarketerID string `query:"marketer_id"`
	LecturerID string `query:"lecturer_id"`
	StudentID  string `query:"student_id"`
	ProgramID  string `query:"program_id"`
	types.MetaQuery
}

type GetPayrollItemsResp struct {
	Items []PayrollItem `json:"items"`
	Meta  types.Meta    `json:"meta"`
}

type UpdatePayrollItemReq struct {
	ID                 string           `params:"id" validate:"required,ulid"`
	UserID             string           `json:"user_id"`
	ProgramMeetings    *uint64          `json:"program_meetings"`
	AcquisitionRights  *uint64          `json:"acquisition_rights"`
	IsMeetingFull      *bool            `json:"is_meeting_full"`
	ForeignLearningFee *decimal.Decimal `json:"foreign_learning_fee"`
	NightLearningFee   *decimal.Decimal `json:"night_learning_fee"`
	Wage               *decimal.Decimal `json:"wage"`
	InitialWage        *decimal.Decimal `json:"initial_wage"`
}

func (u *UpdatePayrollItemReq) Validate() error {
	return nil
}

type DeletePayrollItemReq struct {
	ID     string `params:"id" validate:"required,ulid"`
	UserID string `json:"user_id"`
}
