package entity

type CopyRegistrationsReq struct {
	UserId        string          `json:"user_id" validate:"required,ulid"`
	Registrations []CopyRegisItem `validate:"required,dive"`
}

type CopyRegisItem struct {
	RegisId     string `json:"registration_id" validate:"required,ulid"`
	AllocatedAt string `json:"allocated_at" validate:"required,datetime=2006-01-02"`
	Timezone    string `json:"timezone" validate:"required,timezone"`
}
