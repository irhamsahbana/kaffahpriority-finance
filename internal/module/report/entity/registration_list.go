package entity

import (
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/types"
)

type GetRegistrationsReq struct {
	UserId string `validate:"required,ulid"`

	Q          string `query:"q" validate:"omitempty,min=3"` // search by student name
	PaidAtFrom string `query:"paid_at_from" validate:"omitempty,datetime=2006-01-02"`
	PaidAtTo   string `query:"paid_at_to" validate:"omitempty,datetime=2006-01-02"`
	Timezone   string `query:"timezone" validate:"required,timezone"`

	MarketerId string `query:"marketer_id" validate:"omitempty,ulid"`
	LecturerId string `query:"lecturer_id" validate:"omitempty,ulid"`
	StudentId  string `query:"student_id" validate:"omitempty,ulid"`
	ProgramId  string `query:"program_id" validate:"omitempty,ulid"`

	SortBy   string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at paid_at student_name"`
	SortType string `query:"sort_type" validate:"omitempty,oneof=asc desc"`

	types.MetaQuery
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
	Id                    string       `json:"id" db:"id"`
	ProgramId             string       `json:"program_id" db:"program_id"`
	MarketerId            string       `json:"marketer_id" db:"marketer_id"`
	LecturerId            *string      `json:"lecturer_id" db:"lecturer_id"`
	StudentId             string       `json:"student_id" db:"student_id"`
	StudentIdentifier     string       `json:"student_identifier" db:"student_identifier"`
	ProgramName           string       `json:"program_name" db:"program_name"`
	LecturerName          *string      `json:"lecturer_name" db:"lecturer_name"`
	MarketerName          string       `json:"marketer_name" db:"marketer_name"`
	StudentName           string       `json:"student_name" db:"student_name"`
	MonthlyFee            float64      `json:"monthly_fee" db:"monthly_fee"`
	Students              []AddStudent `json:"additional_students"`
	ProgramFee            float64      `json:"program_fee" db:"program_fee"`
	AdministrationFee     *float64     `json:"administration_fee" db:"administration_fee"`
	FLFee                 *float64     `json:"foreign_learning_fee" db:"foreign_learning_fee"`
	NLFee                 *float64     `json:"night_learning_fee" db:"night_learning_fee"`
	MarketerCommissionFee float64      `json:"marketer_commission_fee" db:"marketer_commission_fee"`
	OverpaymentFee        *float64     `json:"overpayment_fee" db:"overpayment_fee"`
	HRFee                 float64      `json:"hr_fee" db:"hr_fee"`
	HRFeeForMentor        *float64     `json:"hr_fee_for_mentor" db:"hr_fee_for_mentor"`
	HRFeeForHR            *float64     `json:"hr_fee_for_hr" db:"hr_fee_for_hr"`
	MarketerGiftsFee      float64      `json:"marketer_gifts_fee" db:"marketer_gifts_fee"`
	ClosingFeeForOffice   *float64     `json:"closing_fee_for_office" db:"closing_fee_for_office"`
	ClosingFeeForReward   *float64     `json:"closing_fee_for_reward" db:"closing_fee_for_reward"`
	Profit                float64      `json:"profit" db:"profit"`
	Notes                 *string      `json:"notes" db:"notes"`
	PaidAt                string       `json:"paid_at" db:"paid_at"`
	CreatedAt             string       `json:"created_at" db:"created_at"`
	UpdatedAt             string       `json:"updated_at" db:"updated_at"`
}
