package entity

import "codebase-app/pkg/types"

type GetAcquisitionRightsAggregateReq struct {
	UserID string `validate:"ulid"`
	types.MetaQuery

	For               string `query:"for" validate:"required,oneof=academic_manager student_manager"`
	AcademicManagerId string `query:"academic_manager_id" validate:"omitempty,ulid"`
	StudentManagerId  string `query:"student_manager_id" validate:"omitempty,ulid"`
	Month             string `query:"month" validate:"omitempty,datetime=2006-01"`
	Timezone          string `query:"timezone" validate:"required,timezone"`
}

func (r *GetAcquisitionRightsAggregateReq) SetDefault() {
	r.MetaQuery.SetDefault()

	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}
}

type GetAcquisitionRightsAggregateResp struct {
	Items []AcquisitionRightsAggregate `json:"items"`
	Meta  types.Meta                   `json:"meta"`
}

type AcquisitionRightsAggregate struct {
	Month                  string  `json:"month" db:"month"`
	AcademicManagerId      string  `json:"academic_manager_id,omitempty" db:"academic_manager_id"`
	AcademicManagerName    string  `json:"academic_manager_name,omitempty" db:"academic_manager_name"`
	StudentManagerId       string  `json:"student_manager_id,omitempty" db:"student_manager_id"`
	StudentManagerName     string  `json:"student_manager_name,omitempty" db:"student_manager_name"`
	TotalAcquisitionRights float64 `json:"total_acquisition_rights" db:"total_acquisition_rights"`
}
