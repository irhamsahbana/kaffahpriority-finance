package entity

type RegistrationMuliAllocationReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	TemplateId  string   `json:"template_id" validate:"required,ulid"`
	PaidAt      string   `json:"paid_at" validate:"required,datetime=2006-01-02"`
	Allocations []string `json:"allocations" validate:"required,unique_in_slice,dive,datetime=2006-01"`
}

func (r *RegistrationMuliAllocationReq) SetDefault() {
}
