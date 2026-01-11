package entity

type UpdateTemplateCreatedAtBetweenReq struct {
	PrevID   string `json:"prev_id"`
	NextID   string `json:"next_id"`
	TargetID string `json:"target_id" validate:"required"`
	UserID   string `json:"user_id"`
}
