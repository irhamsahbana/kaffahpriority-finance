package entity

type UpdateRegistrationIsPaidReq struct {
	UserID string `json:"user_id" validate:"ulid"`

	ID     string `params:"id" validate:"ulid"`
	IsPaid bool   `json:"is_paid"`
}

type UpdateRegistrationIsPaidResp struct {
	ID     string `json:"id"`
	IsPaid bool   `json:"is_paid"`
}
