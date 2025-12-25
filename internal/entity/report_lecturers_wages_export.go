package entity

type GetExportedLecturersWagesReq struct {
	UserID string `validate:"ulid"`

	Q string `query:"q" validate:"omitempty,min=3"`

	AcademicManagerId          string `query:"academic_manager_id" validate:"omitempty,ulid"`
	LecturerID                 string `query:"lecturer_id" validate:"omitempty,ulid"`
	IsMandatoryFieldsCompleted string `query:"is_mandatory_fields_completed" validate:"omitempty,oneof=true false"`
	Month                      string `query:"month" validate:"required,datetime=2006-01"`
	Timezone                   string `query:"timezone" validate:"required,timezone"`
}

func (r *GetExportedLecturersWagesReq) SetDefault() {
	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}
}

type GetExportedLecturersWagesResp struct {
	Items    []LecturersWageItem `json:"items"`
	FilePath string              `json:"-"`
	FileName string              `json:"-"`
}
