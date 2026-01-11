package entity

type UpdateTemplateCreatedAtBetweenReq struct {
	PrevID   string `json:"prev_id" validate:"required"`
	NextID   string `json:"next_id" validate:"required"`
	TargetID string `json:"target_id" validate:"required"`
	UserID   string `json:"user_id"`
}
