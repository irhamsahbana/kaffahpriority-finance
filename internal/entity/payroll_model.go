package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type PayrollRun struct {
	ID          string           `json:"id" db:"id"`
	PeriodStart time.Time        `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time        `json:"period_end" db:"period_end"`
	Period      string           `json:"period" db:"period"`
	Timezone    string           `json:"timezone" db:"timezone"`
	Status      PayrollRunStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	DeletedAt   *time.Time       `json:"deleted_at" db:"deleted_at"`
}

type PayrollRunStatus string

const (
	PayrollRunStatusDraft    PayrollRunStatus = "draft"
	PayrollRunStatusChecking PayrollRunStatus = "checking"
	PayrollRunStatusPaid     PayrollRunStatus = "paid"
)

type PayrollItem struct {
	ID                  string          `json:"id" db:"id"`
	TemplateID          string          `json:"template_id" db:"template_id"`
	PayrollRunID        string          `json:"payroll_run_id" db:"payroll_run_id"`
	AcademicManagerID   string          `json:"academic_manager_id" db:"academic_manager_id"`
	AcademicManagerName string          `json:"academic_manager_name" db:"academic_manager_name"`
	LecturerID          string          `json:"lecturer_id" db:"lecturer_id"`
	LecturerName        string          `json:"lecturer_name" db:"lecturer_name"`
	StudentID           string          `json:"student_id" db:"student_id"`
	StudentName         string          `json:"student_name" db:"student_name"`
	ProgramID           string          `json:"program_id" db:"program_id"`
	ProgramName         string          `json:"program_name" db:"program_name"`
	MarketerID          string          `json:"marketer_id" db:"marketer_id"`
	MarketerName        string          `json:"marketer_name" db:"marketer_name"`
	ForeignLearningFee  decimal.Decimal `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NightLearningFee    decimal.Decimal `json:"night_learning_fee" db:"night_learning_fee"`
	IsITP               bool            `json:"is_itp" db:"is_itp"`
	ProgramMeetings     uint64          `json:"program_meetings" db:"program_meetings"`
	IsMeetingFull       bool            `json:"is_meeting_full" db:"is_meeting_full"`
	WagePerMeeting      decimal.Decimal `json:"wage_per_meeting" db:"wage_per_meeting"`
	InitialWage         decimal.Decimal `json:"initial_wage" db:"initial_wage"`
	FullWage            decimal.Decimal `json:"full_wage" db:"full_wage"`
	Wage                decimal.Decimal `json:"wage" db:"wage"`
	AcquisitionRights   uint64          `json:"acquisition_rights" db:"acquisition_rights"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time      `json:"deleted_at" db:"deleted_at"`

	AdditionalStudents []PayrollItemAdditionalStudent `json:"additional_students" db:"-"`
}

type PayrollItemAdditionalStudent struct {
	ID            string     `json:"id" db:"id"`
	PayrollItemID string     `json:"payroll_item_id" db:"payroll_item_id"`
	StudentID     *string    `json:"student_id" db:"student_id"`
	Name          string     `json:"name" db:"name"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at" db:"deleted_at"`
}
