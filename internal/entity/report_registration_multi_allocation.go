package entity

type RegistrationMuliAllocationReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	TemplateId  string   `json:"template_id" validate:"required,ulid"`
	PaidAt      string   `json:"paid_at" validate:"required,datetime=2006-01-02"`
	PaidAtTime  string   `json:"paid_at_time" validate:"required,datetime=15:04:05"`
	Allocations []string `json:"allocations" validate:"required,unique_in_slice,dive,datetime=2006-01"`
}

func (r *RegistrationMuliAllocationReq) SetDefault() {
}
