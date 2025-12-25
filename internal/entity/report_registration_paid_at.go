package entity

type UpdateRegisPaidAtReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	PaidAt          string   `json:"paid_at" validate:"required,datetime=2006-01-02"`
	PaidAtTime      string   `json:"paid_at_time" validate:"required,datetime=15:04:05"`
	AllocatedAt     string   `json:"allocated_at" validate:"required,datetime=2006-01-02"`
	AllocatedAtTime string   `json:"allocated_at_time" validate:"required,datetime=15:04:05"`
	RegistrationIds []string `json:"registration_ids" validate:"required,min=1,dive,ulid"`
}
