package entity

import (
	"codebase-app/pkg/types"
	"io"

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

type GetPayrollItemsPeriodicallyReq struct {
	Period            string `query:"period"`
	Q                 string `query:"q"`
	MarketerID        string `query:"marketer_id"`
	LecturerID        string `query:"lecturer_id"`
	StudentID         string `query:"student_id"`
	ProgramID         string `query:"program_id"`
	AcademicManagerID string `query:"academic_manager_id"`
	types.MetaQuery
}

func (r *GetPayrollItemsPeriodicallyReq) SetDefault() {
	r.MetaQuery.SetDefault()
}

type PayrollItemPeriodically struct {
	LecturerID             string          `json:"lecturer_id" db:"lecturer_id"`
	LecturerName           string          `json:"lecturer_name" db:"lecturer_name"`
	AcademicManagerID      string          `json:"academic_manager_id" db:"academic_manager_id"`
	AcademicManagerName    string          `json:"academic_manager_name" db:"academic_manager_name"`
	Period                 string          `json:"period" db:"period"`
	TotalRealFee           decimal.Decimal `json:"total_real_fee" db:"total_real_fee"`
	TotalAcquisitionRights uint64          `json:"total_acquisition_rights" db:"total_acquisition_rights"`
}

type GetPayrollItemsPeriodicallyResp struct {
	Items []PayrollItemPeriodically `json:"items"`
	Meta  types.Meta                `json:"meta"`
}

type UpdatePayrollItemReq struct {
	ID                 string           `params:"id" json:"id" validate:"required,ulid"`
	UserID             string           `json:"user_id"`
	ProgramMeetings    *uint64          `json:"program_meetings"`
	AcquisitionRights  *uint64          `json:"acquisition_rights"`
	IsMeetingFull      *bool            `json:"is_meeting_full"`
	ForeignLearningFee *decimal.Decimal `json:"foreign_learning_fee"`
	NightLearningFee   *decimal.Decimal `json:"night_learning_fee"`
	Wage               *decimal.Decimal `json:"wage"`
	InitialWage        *decimal.Decimal `json:"initial_wage"`
}

func (_ *UpdatePayrollItemReq) Validate() error {
	return nil
}

type BulkUpdatePayrollItemReq struct {
	UserID string                 `json:"user_id"`
	Data   []UpdatePayrollItemReq `json:"data" validate:"required,dive"`
}

type DeletePayrollItemReq struct {
	ID     string `params:"id" validate:"required,ulid"`
	UserID string `json:"user_id"`
}

type ExportPayrollItemsPeriodicallyReq struct {
	Period string `query:"period" validate:"required,datetime=2006-01"`
}

type ExportPayrollItemsPeriodicallyResp struct {
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type ImportPayrollItemsReq struct {
	UserID   string    `json:"user_id"`
	File     io.Reader `json:"-"`
	FileName string    `json:"file_name"`
	Items    []ImportedPayrollItems
}

type ImportPayrollItemsResp struct {
	TotalProcessed int `json:"total_processed"`
	TotalUpdated   int `json:"total_updated"`
}

type ImportedPayrollItems struct {
	ID                 string           `json:"id" db:"id"`
	ProgramMeetings    int              `json:"program_meetings" db:"program_meetings"`
	InitialWage        decimal.Decimal  `json:"initial_wage" db:"initial_wage"`
	IsMeetingFull      bool             `json:"is_meeting_full" db:"is_meeting_full"`
	ForeignLearningFee *decimal.Decimal `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NightLearningFee   *decimal.Decimal `json:"night_learning_fee" db:"night_learning_fee"`
	AcquisitionRights  *uint64          `json:"acquisition_rights" db:"acquisition_rights"`
	Wage               decimal.Decimal  `json:"wage" db:"wage"`
	FullWage           decimal.Decimal  `json:"full_wage" db:"full_wage"`
}
