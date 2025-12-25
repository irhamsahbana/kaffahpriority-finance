package entity

import (
	"codebase-app/pkg/types"
	"time"

	"github.com/shopspring/decimal"
)

type GetLecturersWagesReq struct {
	UserID string `validate:"ulid"`
	types.MetaQuery

	Q string `query:"q" validate:"omitempty,min=3"`

	AcademicManagerID          string `query:"academic_manager_id" validate:"omitempty,ulid"`
	LecturerID                 string `query:"lecturer_id" validate:"omitempty,ulid"`
	IsMandatoryFieldsCompleted string `query:"is_mandatory_fields_completed" validate:"omitempty,oneof=true false"`
	Month                      string `query:"month" validate:"omitempty,datetime=2006-01"`
	Timezone                   string `query:"timezone" validate:"required,timezone"`
}

func (r *GetLecturersWagesReq) SetDefault() {
	r.MetaQuery.SetDefault()

	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}

	if r.Month == "" {
		// get current month with timezone
		loc, _ := time.LoadLocation(r.Timezone)
		r.Month = time.Now().In(loc).Format("2006-01")
	}
}

type GetLecturersWagesResp struct {
	Items []LecturersWageItem `json:"items"`
	Meta  types.Meta          `json:"meta"`
}

type LecturersWageItem struct {
	RegistrationID       string           `json:"registration_id" db:"registration_id"`
	AcademicManagerName  *string          `json:"academic_manager_name" db:"academic_manager_name"`
	LecturerName         *string          `json:"lecturer_name" db:"lecturer_name"`
	StudentName          string           `json:"student_name" db:"student_name"`
	ProgramName          string           `json:"program_name" db:"program_name"`
	StudentManagerName   string           `json:"student_manager_name" db:"student_manager_name"`
	MarketerName         string           `json:"marketer_name" db:"marketer_name"`
	AccquisitionRights   int              `json:"acquisition_rights" db:"acquisition_rights"`
	FL                   *decimal.Decimal `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NL                   *decimal.Decimal `json:"night_learning_fee" db:"night_learning_fee"`
	IsITP                bool             `json:"is_itp" db:"is_itp"`
	ProgramMeetings      int              `json:"program_meetings" db:"program_meetings"`
	ProgramFeePerMeeting decimal.Decimal  `json:"program_fee_per_meeting" db:"program_fee_per_meeting"`
	IsFullFee            bool             `json:"is_full_fee" db:"is_full_fee"`                       //
	FullFee              decimal.Decimal  `json:"full_fee" db:"full_fee"`                             // ujrah full
	InitialFee           decimal.Decimal  `json:"initial_fee" db:"initial_fee"`                       // ujrah awal
	RealFee              decimal.Decimal  `json:"real_fee" db:"real_fee"`                             // ujrah real
	MentorDetailFeeUsed  *decimal.Decimal `json:"mentor_detail_fee_used" db:"mentor_detail_fee_used"` // wage for mentor / keep gaji
	AllocatedAt          *string          `json:"allocated_at" db:"allocated_at"`
	Notes                *string          `json:"notes" db:"notes"`
	TemplateId           string           `json:"template_id" db:"template_id"`
	HasPayment           bool             `json:"has_payment" db:"has_payment"`
}

// aggregate version

type GetLecturersWagesAggregateReq struct {
	UserID string `validate:"ulid"`
	types.MetaQuery

	AcademicManagerID string `query:"academic_manager_id" validate:"omitempty,ulid"`
	LecturerID        string `query:"lecturer_id" validate:"omitempty,ulid"`
	Month             string `query:"month" validate:"required,datetime=2006-01"`
	Timezone          string `query:"timezone" validate:"required,timezone"`
}

func (r *GetLecturersWagesAggregateReq) SetDefault() {
	r.MetaQuery.SetDefault()

	r.Timezone = "Asia/Makassar"

	if r.Month == "" {
		loc, _ := time.LoadLocation(r.Timezone)
		r.Month = time.Now().In(loc).Format("2006-01")
	}
}

type LecturersWageAggregateResp struct {
	Items []LecturersWageAggregateItem `json:"items"`
	Meta  types.Meta                   `json:"meta"`
}

type LecturersWageAggregateItem struct {
	Month                  string          `json:"month" db:"month"`
	LecturerID             string          `json:"lecturer_id" db:"lecturer_id"`
	AcademicManagerID      string          `json:"academic_manager_id" db:"academic_manager_id"`
	LecturerName           string          `json:"lecturer_name" db:"lecturer_name"`
	AcademicManagerName    string          `json:"academic_manager_name" db:"academic_manager_name"`
	TotalRealFee           decimal.Decimal `json:"total_real_fee" db:"total_real_fee"`
	TotalAcquisitionRights int             `json:"total_acquisition_rights" db:"total_acquisition_rights"`
}

// yearly aggregate version

type GetLecturersWagesAggregateYearlyReq struct {
	UserID string `validate:"ulid"`
	types.MetaQuery

	AcademicManagerID string `query:"academic_manager_id" validate:"omitempty,ulid"`
	LecturerID        string `query:"lecturer_id" validate:"omitempty,ulid"`
	Year              int    `query:"year" validate:"required,min=1900,max=2100"`
	Timezone          string `query:"timezone" validate:"required,timezone"`
}

func (r *GetLecturersWagesAggregateYearlyReq) SetDefault() {
	r.MetaQuery.SetDefault()

	r.Timezone = "Asia/Makassar"

	if r.Year == 0 {
		loc, _ := time.LoadLocation(r.Timezone)
		r.Year = time.Now().In(loc).Year()
	}
}

type LecturersWageAggregateYearlyResp struct {
	Items []LecturersWageAggregateYearlyItem `json:"items"`
	Meta  types.Meta                         `json:"meta"`
}

type LecturersWageAggregateYearlyItem struct {
	LecturerID          string                   `json:"lecturer_id" db:"lecturer_id"`
	LecturerName        string                   `json:"lecturer_name" db:"lecturer_name"`
	AcademicManagerName string                   `json:"academic_manager_name" db:"academic_manager_name"`
	Year                int                      `json:"year" db:"year"`
	Months              []LecturersWageMonthItem `json:"months"`
}

type LecturersWageMonthItem struct {
	Month                  string          `json:"month"`
	UsedAmount             decimal.Decimal `json:"used_amount"`
	Notes                  *string         `json:"notes"`
	TotalAcquisitionRights int             `json:"total_acquisition_rights"`
}
