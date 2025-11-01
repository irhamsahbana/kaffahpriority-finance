package entity

type RegistrationsMarkAsUsedReq struct {
	UserID         string `json:"user_id" validate:"required,ulid"`
	AllocatedMonth string `json:"allocated_month" validate:"required,datetime=2006-01"`
}
