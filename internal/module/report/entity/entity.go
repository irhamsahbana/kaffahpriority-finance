package entity

type UpdateTemplateReq struct {
	Id string `params:"id" validate:"ulid"`

	CreateTemplateReq
}

type UpdateTemplateResp struct {
	Id string `json:"id"`
}
