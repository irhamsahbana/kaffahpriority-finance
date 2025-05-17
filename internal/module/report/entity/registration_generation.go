package entity

type GenerateRegistrationsReq struct {
	UserId   string `json:"user_id" validate:"required,ulid"`
	Timezone string `query:"timezone" validate:"required,timezone"`
}

func (r *GenerateRegistrationsReq) SetDefault() {
	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}
}
