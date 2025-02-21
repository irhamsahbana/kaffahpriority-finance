package entity

import (
	"codebase-app/pkg/types"
	"time"

	"github.com/shopspring/decimal"
)

type GetRegistrationListPerLecturerReq struct {
	UserId string `json:"user_id"`
	types.MetaQuery
	Year int    `query:"year"`
	Tz   string `query:"timezone"`
}

func (r *GetRegistrationListPerLecturerReq) SetDefault() {
	r.MetaQuery.SetDefault()

	if r.Tz == "" {
		r.Tz = "Asia/Makassar"
	}

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
	LecturerId    *string                               `json:"lecturer_id" db:"lecturer_id"`
	StudentId     string                                `json:"student_id" db:"student_id"`
	ProgramId     string                                `json:"program_id" db:"program_id"`
	LecturerName  *string                               `json:"lecturer_name" db:"lecturer_name"`
	StudentName   string                                `json:"student_name" db:"student_name"`
	ProgramName   string                                `json:"program_name" db:"program_name"`
	Year          int                                   `json:"year"`
	IsFL          bool                                  `json:"is_fl" db:"is_fl"`
	IsNL          bool                                  `json:"is_nl" db:"is_nl"`
	Registrations []RegistrationListPerLecturerPerMonth `json:"registrations"`
}

type RegistrationListPerLecturerPerMonth struct {
	RegistrationId *string          `json:"registration_id" db:"registration_id"`
	Month          string           `json:"month" db:"month"` // indonesia month
	MonthNum       int              `json:"month_num" db:"month_num"`
	UsedAmount     *decimal.Decimal `json:"used_amount" db:"used_amount"`
	HRFeeLecturer  *decimal.Decimal `json:"hr_fee_for_lecturer" db:"hr_fee_for_lecturer"`
	IsUsed         *bool            `json:"is_used" db:"is_used"`
	Notes          *string          `json:"notes" db:"notes"`

	ProgramId  string           `json:"program_id" db:"program_id"`
	LecturerId *string          `json:"lecturer_id" db:"lecturer_id"`
	StudentId  string           `json:"student_id" db:"student_id"`
	FL         *decimal.Decimal `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NL         *decimal.Decimal `json:"night_learning_fee" db:"night_learning_fee"`
}
