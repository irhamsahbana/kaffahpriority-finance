package entity

import (
	"codebase-app/pkg/types"
	"time"

	"github.com/shopspring/decimal"
)

type GetRegistrationListPerLecturerReq struct {
	UserID string `json:"user_id"`
	types.MetaQuery

	Q                 string `query:"q"`
	LecturerID        string `query:"lecturer_id"`
	StudentID         string `query:"student_id"`
	AcademicManagerID string `query:"academic_manager_id"`
	Year              int    `query:"year"`
	Tz                string `query:"timezone"`
	IsStarted         string `query:"is_started"`
}

func (r *GetRegistrationListPerLecturerReq) SetDefault() {
	r.MetaQuery.SetDefault()

	r.Tz = "Asia/Makassar"

	if r.Year < 1 {
		year := time.Now().In(time.FixedZone(r.Tz, 0)).Year()
		r.Year = year
	}

}

type GetRegistrationListPerLecturerResp struct {
	Items []RegistrationListPerLecturer `json:"items"`
	Meta  types.Meta                    `json:"meta"`
}

type RegistrationListPerLecturer struct {
	TemplateID          string                                `json:"template_id" db:"template_id"`
	AcademicManagerID   *string                               `json:"academic_manager_id" db:"academic_manager_id"`
	LecturerID          *string                               `json:"lecturer_id" db:"lecturer_id"`
	StudentID           string                                `json:"student_id" db:"student_id"`
	ProgramID           string                                `json:"program_id" db:"program_id"`
	AcademicManagerName *string                               `json:"academic_manager_name" db:"academic_manager_name"`
	LecturerName        *string                               `json:"lecturer_name" db:"lecturer_name"`
	StudentName         string                                `json:"student_name" db:"student_name"`
	ProgramName         string                                `json:"program_name" db:"program_name"`
	Year                int                                   `json:"year"`
	IsFL                bool                                  `json:"is_fl" db:"is_fl"`
	IsNL                bool                                  `json:"is_nl" db:"is_nl"`
	IsITP               bool                                  `json:"is_itp" db:"is_itp"`
	IsStarted           bool                                  `json:"is_started" db:"is_started"`
	Registrations       []RegistrationListPerLecturerPerMonth `json:"registrations"`
}

type RegistrationListPerLecturerPerMonth struct {
	TemplateID     string           `json:"template_id" db:"template_id"`
	RegistrationID *string          `json:"registration_id" db:"registration_id"`
	Month          string           `json:"month" db:"month"` // indonesia month
	MonthNum       int              `json:"month_num" db:"month_num"`
	UsedAmount     *decimal.Decimal `json:"used_amount" db:"used_amount"`
	HRFeeLecturer  *decimal.Decimal `json:"hr_fee_for_lecturer" db:"hr_fee_for_lecturer"`
	IsUsed         *bool            `json:"is_used" db:"is_used"`
	Notes          *string          `json:"notes" db:"notes"`

	ProgramID  string           `json:"program_id" db:"program_id"`
	LecturerID *string          `json:"lecturer_id" db:"lecturer_id"`
	StudentID  string           `json:"student_id" db:"student_id"`
	FL         *decimal.Decimal `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NL         *decimal.Decimal `json:"night_learning_fee" db:"night_learning_fee"`
	IsITP      bool             `json:"is_itp" db:"is_itp"`
	IsStarted  bool             `json:"is_started" db:"is_started"`
}
