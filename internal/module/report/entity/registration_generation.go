package entity

type GenerateRegistrationsReq struct {
	UserID   string `json:"user_id" validate:"required,ulid"`
	Timezone string `validate:"required,timezone"`
}

func (r *GenerateRegistrationsReq) SetDefault() {
	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}
}
