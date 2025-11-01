package entity

type RegistrationsMarkAsUsedReq struct {
	UserID          string   `json:"user_id" validate:"required,ulid"`
	RegistrationIds []string `json:"registration_ids" validate:"required,min=1,dive,ulid"`
}
