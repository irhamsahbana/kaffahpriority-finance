package entity

import (
	"codebase-app/pkg/errmsg"

	"github.com/shopspring/decimal"
)

type UseHRfeeForLecturerReq struct {
	UserId string `json:"user_id" validate:"ulid"`

	RegistrationId string           `params:"registration_id" validate:"ulid"`
	UsedAmount     *decimal.Decimal `json:"used_amount"`
	Notes          *string          `json:"notes" validate:"omitempty,max=255"`
}

func (r *UseHRfeeForLecturerReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	if r.UsedAmount != nil && r.UsedAmount.LessThanOrEqual(decimal.Zero) {
		err.Add("used_amount", "used amount must be greater than 0")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}
