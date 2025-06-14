package entity

type UpdateRegisPaidAtReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	PaidAt          string   `json:"paid_at" validate:"required,datetime=2006-01-02"`
	RegistrationIds []string `json:"registration_ids" validate:"required,min=1,dive,ulid"`
}
