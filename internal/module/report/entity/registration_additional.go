package entity

type CreateAdditionalRegistrationReq struct {
	UserID           string  `json:"user_id" validate:"required,ulid"`
	ParentID         string  `params:"id" validate:"required,ulid"`
	Category         string  `json:"category" validate:"required,oneof=general additional shortfall"`
	NotesForCategory *string `json:"notes_for_category" validate:"omitempty,max=255"`

	PaidAt     string `json:"paid_at" validate:"required,datetime=2006-01-02"`
	PaidAtTime string `json:"paid_at_time" validate:"required,datetime=15:04:05"`

	HrFee                 float64 `json:"hr_fee" validate:"number,gte=0"`
	MarketerCommissionFee float64 `json:"marketer_commission_fee" validate:"number,gte=0"`
	MarketerGiftsFee      float64 `json:"marketer_gifts_fee" validate:"number,gte=0"`

	ClosingFeeForOffice float64 `json:"closing_fee_for_office" validate:"number,gte=0"`
	ClosingFeeForReward float64 `json:"closing_fee_for_reward" validate:"number,gte=0"`
}

type CreateAdditionalRegistrationResp struct {
	ID string `json:"id"`
}
