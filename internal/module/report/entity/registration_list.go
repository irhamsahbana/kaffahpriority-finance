package entity

import (
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/types"
)

type GetRegistrationsReq struct {
	UserID string `validate:"required,ulid"`

	Q              string `query:"q" validate:"omitempty,min=3"` // search by student name
	PaidAtFrom     string `query:"paid_at_from" validate:"omitempty,datetime=2006-01-02"`
	PaidAtTo       string `query:"paid_at_to" validate:"omitempty,datetime=2006-01-02"`
	AllocatedMonth string `query:"allocated_month" validate:"omitempty,datetime=2006-01"`
	Timezone       string `query:"timezone" validate:"required,timezone"`
	IsPaid         string `query:"is_paid" validate:"omitempty,oneof=true false"`
	IsStarted      string `query:"is_started" validate:"omitempty,oneof=true false"`

	MarketerID string `query:"marketer_id" validate:"omitempty,ulid"`
	LecturerID string `query:"lecturer_id" validate:"omitempty,ulid"`
	StudentID  string `query:"student_id" validate:"omitempty,ulid"`
	ProgramID  string `query:"program_id" validate:"omitempty,ulid"`

	// mentor_detail_fee_used
	IsLecturerFeeUsed          string `query:"is_lecturer_fee_used" validate:"omitempty,oneof=true false"`
	IsMandatoryFieldsCompleted string `query:"is_mandatory_fields_completed" validate:"omitempty,oneof=true false"`
	MentorFeeAllocationStatus  string `query:"mentor_fee_allocation_status" validate:"omitempty,oneof=all full partial none"`

	SortBy   string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at paid_at student_name"`
	SortType string `query:"sort_type" validate:"omitempty,oneof=asc desc"`

	types.MetaQuery

	GetUnusedRegistrationsReq
}

type GetUnusedRegistrationsReq struct {
	ExcludeAllocatedMonth string `query:"exclude_allocated_month" validate:"omitempty,datetime=2006-01"`
}

func (r *GetRegistrationsReq) SetDefault() {
	r.MetaQuery.SetDefault()

	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}

	if r.SortBy == "" {
		r.SortBy = "paid_at"
	}

	if r.SortType == "" {
		r.SortType = "desc"
	}

	if r.MentorFeeAllocationStatus == "" {
		r.MentorFeeAllocationStatus = "all"
	}
}

func (r *GetRegistrationsReq) Validate() error {
	err := errmsg.NewCustomErrors(400)
	if (r.PaidAtFrom != "" || r.PaidAtTo != "") && (r.PaidAtFrom == "" || r.PaidAtTo == "") { // if one of them is empty
		err.Add("paid_at_from", "batas bawah tanggal pembayaran harus diisi")
		err.Add("paid_at_to", "batas atas tanggal pembayaran harus diisi")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type GetRegistrationsResp struct {
	Items []RegisItem `json:"items"`
	Meta  types.Meta  `json:"meta"`
}

type RegisItem struct {
	ID                         string       `json:"id" db:"id"`
	TemplateID                 string       `json:"template_id" db:"template_id"`
	ProgramID                  string       `json:"program_id" db:"program_id"`
	MarketerID                 string       `json:"marketer_id" db:"marketer_id"`
	StudentManagerID           string       `json:"student_manager_id" db:"student_manager_id"`
	LecturerID                 *string      `json:"lecturer_id" db:"lecturer_id"`
	AcademicManagerID          *string      `json:"academic_manager_id" db:"academic_manager_id"`
	StudentID                  string       `json:"student_id" db:"student_id"`
	StudentIdentifier          string       `json:"student_identifier" db:"student_identifier"`
	IsMandatoryFieldsCompleted bool         `json:"is_mandatory_fields_completed" db:"is_mandatory_fields_completed"`
	ProgramName                string       `json:"program_name" db:"program_name"`
	LecturerName               *string      `json:"lecturer_name" db:"lecturer_name"`
	AcademicManagerName        *string      `json:"academic_manager_name" db:"academic_manager_name"`
	MarketerName               string       `json:"marketer_name" db:"marketer_name"`
	StudentManagerName         string       `json:"student_manager_name" db:"student_manager_name"`
	StudentName                string       `json:"student_name" db:"student_name"`
	MonthlyFee                 float64      `json:"monthly_fee" db:"monthly_fee"`
	Students                   []AddStudent `json:"additional_students"`
	ProgramFee                 float64      `json:"program_fee" db:"program_fee"`
	AdministrationFee          *float64     `json:"administration_fee" db:"administration_fee"`
	FLFee                      *float64     `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NLFee                      *float64     `json:"night_learning_fee" db:"night_learning_fee"`
	IsITP                      bool         `json:"is_itp" db:"is_itp"`
	AcquisitionRights          int          `json:"acquisition_rights" db:"acquisition_rights"`
	MarketerCommissionFee      float64      `json:"marketer_commission_fee" db:"marketer_commission_fee"`
	OverpaymentFee             *float64     `json:"overpayment_fee" db:"overpayment_fee"`
	HRFee                      float64      `json:"hr_fee" db:"hr_fee"`
	HRFeeForMentor             *float64     `json:"hr_fee_for_mentor" db:"hr_fee_for_mentor"`
	HRFeeForHR                 *float64     `json:"hr_fee_for_hr" db:"hr_fee_for_hr"`
	HRFeeForMentorRemaining    *float64     `json:"hr_fee_for_mentor_remaining" db:"hr_fee_for_mentor_remaining"`
	HRFeeForMentorStatus       *string      `json:"hr_fee_for_mentor_status" db:"hr_fee_for_mentor_status"`
	IsMentorDetailFeeUsed      bool         `json:"is_mentor_detail_fee_used" db:"is_mentor_detail_fee_used"`
	MarketerGiftsFee           float64      `json:"marketer_gifts_fee" db:"marketer_gifts_fee"`
	ClosingFeeForOffice        *float64     `json:"closing_fee_for_office" db:"closing_fee_for_office"`
	ClosingFeeForReward        *float64     `json:"closing_fee_for_reward" db:"closing_fee_for_reward"`
	Profit                     float64      `json:"profit" db:"profit"`
	Notes                      *string      `json:"notes" db:"notes"`
	Batch                      *string      `json:"batch" db:"batch"`
	Category                   string       `json:"category" db:"category"`
	IsPaid                     bool         `json:"is_paid" db:"is_paid"`
	IsStarted                  bool         `json:"is_started" db:"is_started"`
	PaidAt                     string       `json:"paid_at" db:"paid_at"`
	CreatedAt                  string       `json:"created_at" db:"created_at"`
	UpdatedAt                  string       `json:"updated_at" db:"updated_at"`
	AllocatedAt                *string      `json:"allocated_at" db:"allocated_at"`

	// internal use only
	IsUnused bool `json:"-"`
}

type GetExportedRegistrationsReq struct {
	UserID string `validate:"required,ulid"`

	Q          string `query:"q" validate:"omitempty,min=3"` // search by student name
	PaidAtFrom string `query:"paid_at_from" validate:"datetime=2006-01-02"`
	PaidAtTo   string `query:"paid_at_to" validate:"datetime=2006-01-02"`
	Timezone   string `query:"timezone" validate:"required,timezone"`

	MarketerID string `query:"marketer_id" validate:"omitempty,ulid"`
	LecturerID string `query:"lecturer_id" validate:"omitempty,ulid"`
	StudentID  string `query:"student_id" validate:"omitempty,ulid"`
	ProgramID  string `query:"program_id" validate:"omitempty,ulid"`

	// mentor_detail_fee_used
	IsLecturerFeeUsed          string `query:"is_lecturer_fee_used" validate:"omitempty,oneof=true false"`
	IsMandatoryFieldsCompleted string `query:"is_mandatory_fields_completed" validate:"omitempty,oneof=true false"`
	MentorFeeAllocationStatus  string `query:"mentor_fee_allocation_status" validate:"omitempty,oneof=all full partial none"`

	SortBy   string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at paid_at student_name"`
	SortType string `query:"sort_type" validate:"omitempty,oneof=asc desc"`
}

func (r *GetExportedRegistrationsReq) SetDefault() {

	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}

	if r.SortBy == "" {
		r.SortBy = "paid_at"
	}

	if r.SortType == "" {
		r.SortType = "desc"
	}

	if r.MentorFeeAllocationStatus == "" {
		r.MentorFeeAllocationStatus = "all"
	}
}

func (r *GetExportedRegistrationsReq) Validate() error {
	err := errmsg.NewCustomErrors(400)
	if (r.PaidAtFrom != "" || r.PaidAtTo != "") && (r.PaidAtFrom == "" || r.PaidAtTo == "") { // if one of them is empty
		err.Add("paid_at_from", "batas bawah tanggal pembayaran harus diisi")
		err.Add("paid_at_to", "batas atas tanggal pembayaran harus diisi")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type GetExportedRegistrationsResp struct {
	Items   []RegisItem       `json:"items"`
	Summary *GetSummariesResp `json:"summary"`

	// internal use only
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type GetExportedRegistrationsForCFO2MonthlyReq struct {
	UserID string `validate:"required,ulid"`

	PaidAtFrom string `query:"paid_at_from" validate:"datetime=2006-01-02"`
	PaidAtTo   string `query:"paid_at_to" validate:"datetime=2006-01-02"`
	Timezone   string `query:"timezone" validate:"required,timezone"`
}

func (r *GetExportedRegistrationsForCFO2MonthlyReq) SetDefault() {
	r.Timezone = "Asia/Makassar"
}

func (r *GetExportedRegistrationsForCFO2MonthlyReq) Validate() error {
	err := errmsg.NewCustomErrors(400)
	if r.PaidAtFrom == "" || r.PaidAtTo == "" { // if one of them is empty
		err.Add("paid_at_from", "batas bawah tanggal pembayaran harus diisi")
		err.Add("paid_at_to", "batas atas tanggal pembayaran harus diisi")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type GetExportedRegistrationsForCFO2MonthlyResp struct {
	Items    []RegisItem `json:"items"`
	TotalITP int64       `json:"total_itp"`

	// internal use only
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type GetExportedRegistrationsForCFO2YearlyReq struct {
	UserID string `validate:"required,ulid"`

	PaidAtYear string `query:"paid_at_year" validate:"required,datetime=2006"`
	Timezone   string `query:"timezone" validate:"required,timezone"`
}

type GetExportedRegistrationsForCFO2YearlyResp struct {
	Items []RegistrationYearlyRow `json:"items"`

	// internal use only
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type RegistrationYearlyRow struct {
	AcademicManagerID   string                    `db:"academic_manager_id" json:"academic_manager_id"`
	AcademicManagerName string                    `db:"academic_manager_name" json:"academic_manager_name"`
	LecturerID          string                    `db:"lecturer_id" json:"lecturer_id"`
	LecturerName        string                    `db:"lecturer_name" json:"lecturer_name"`
	StudentID           string                    `db:"student_id" json:"student_id"`
	StudentName         string                    `db:"student_name" json:"student_name"`
	ProgramID           string                    `db:"program_id" json:"program_id"`
	ProgramName         string                    `db:"program_name" json:"program_name"`
	MarketerID          string                    `db:"marketer_id" json:"marketer_id"`
	MarketerName        string                    `db:"marketer_name" json:"marketer_name"`
	Months              []RegistrationYearlyMonth `json:"months"`
}

type RegistrationYearlyMonth struct {
	RegistrationID string   `db:"registration_id" json:"registration_id"`
	HRFeeForMentor *float64 `db:"hr_fee_for_mentor" json:"hr_fee_for_mentor"`
	PaidAt         string   `db:"paid_at" json:"paid_at"`
	PaidAtMonth    int      `db:"paid_at_month" json:"paid_at_month"`
	Notes          *string  `db:"notes" json:"notes"`
}

type GetExportedRegistrationsForWageRecapMonthlyReq struct {
	UserID string `validate:"required,ulid"`

	Month    string `query:"month" validate:"required,datetime=2006-01"`
	Timezone string `query:"timezone" validate:"required,timezone"`
}

func (r *GetExportedRegistrationsForWageRecapMonthlyReq) SetDefault() {
	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}
}

type GetExportedRegistrationsForWageRecapMonthlyResp struct {
	Items []WageRecapAcademicManager `json:"items"`

	// internal use only
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type WageRecapAcademicManager struct {
	ID    string              `db:"id" json:"id"`
	Name  string              `db:"name" json:"name"`
	Items []WageRecapLecturer `json:"items"`
}

type WageRecapLecturer struct {
	ID    string                  `db:"id" json:"id"`
	Name  string                  `db:"name" json:"name"`
	Items []WageRecapRegistration `json:"items"`
}

type WageRecapRegistration struct {
	ID           string `db:"id" json:"id"`
	StudentID    string `db:"student_id" json:"student_id"`
	ProgramID    string `db:"program_id" json:"program_id"`
	LecturerID   string `db:"lecturer_id" json:"lecturer_id"`
	StudentName  string `db:"student_name" json:"student_name"`
	ProgramName  string `db:"program_name" json:"program_name"`
	MarketerName string `db:"marketer_name" json:"marketer_name"`

	IsFL  bool `db:"is_fl" json:"is_fl"`
	IsNL  bool `db:"is_nl" json:"is_nl"`
	IsITP bool `db:"is_itp" json:"is_itp"`

	Data *WageRecapRegistrationData `db:"data" json:"data"`
}

type WageRecapRegistrationData struct {
	ProgramMeetings      int64    `db:"program_meetings" json:"program_meetings"`
	ProgramFeePerMeeting float64  `db:"program_fee_per_meeting" json:"program_fee_per_meeting"`
	FullFee              float64  `db:"full_fee" json:"full_fee"`
	IsFullFee            bool     `db:"is_full_fee" json:"is_full_fee"`
	InitialFee           *float64 `db:"initial_fee" json:"initial_fee"`
	FL                   *float64 `db:"foreign_learning_fee" json:"foreign_learning_fee"`
	NL                   *float64 `db:"night_learning_fee" json:"night_learning_fee"`
	RealFee              float64  `db:"real_fee" json:"real_fee"`
	MentorDetailFeeUsed  *float64 `db:"mentor_detail_fee_used" json:"mentor_detail_fee_used"`
	AcquisitionRights    int64    `db:"acquisition_rights" json:"acquisition_rights"`
	Notes                *string  `db:"notes" json:"notes"`
}
