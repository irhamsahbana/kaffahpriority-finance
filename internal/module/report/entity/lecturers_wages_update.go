package entity

import (
	"codebase-app/pkg/errmsg"

	"github.com/LukaGiorgadze/gonull"
	"github.com/shopspring/decimal"
)

type UpdateLecturersWageReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	RegistrationId string `json:"registration_id" validate:"required,ulid"`

	ProgramMeetings gonull.Nullable[int]             `json:"program_meetings"`
	InitialFee      gonull.Nullable[decimal.Decimal] `json:"initial_fee"`
	FL              gonull.Nullable[decimal.Decimal] `json:"foreign_learning_fee"`
	NL              gonull.Nullable[decimal.Decimal] `json:"night_learning_fee"`
	IsFullFee       gonull.Nullable[bool]            `json:"is_full_fee"`
}

func (r *UpdateLecturersWageReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	if r.ProgramMeetings.Present && r.ProgramMeetings.Val < 0 {
		err.Add("program_meetings", "program_meetings must be greater than or equal to 0")
	}

	if r.InitialFee.Present && r.InitialFee.Val.LessThan(decimal.Zero) {
		err.Add("initial_fee", "initial_fee must be greater than or equal to 0")
	}

	if r.FL.Present && r.FL.Val.LessThan(decimal.Zero) {
		err.Add("foreign_learning_fee", "foreign_learning_fee must be greater than or equal to 0")
	}

	if r.NL.Present && r.NL.Val.LessThan(decimal.Zero) {
		err.Add("night_learning_fee", "night_learning_fee must be greater than or equal to 0")
	}

	if r.IsFullFee.Present && !r.IsFullFee.Valid {
		err.Add("is_full_fee", "is_full_fee must be a boolean")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type UpdateLecturersWageResp struct {
	RegistrationId string `json:"registration_id"`

	ProgramMeetings gonull.Nullable[int]             `json:"program_meetings"`
	InitialFee      gonull.Nullable[decimal.Decimal] `json:"initial_fee"`
	FL              gonull.Nullable[decimal.Decimal] `json:"foreign_learning_fee"`
	NL              gonull.Nullable[decimal.Decimal] `json:"night_learning_fee"`
	IsFullFee       gonull.Nullable[bool]            `json:"is_full_fee"`
}
