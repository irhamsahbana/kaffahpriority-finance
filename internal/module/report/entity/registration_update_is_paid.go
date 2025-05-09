package entity

type UpdateRegistrationIsPaidReq struct {
	UserId string `json:"user_id" validate:"ulid"`

	Id     string `params:"id" validate:"ulid"`
	IsPaid bool   `json:"is_paid" validate:"required"`
}

type UpdateRegistrationIsPaidResp struct {
	Id     string `json:"id"`
	IsPaid bool   `json:"is_paid"`
}
