package entity

import "codebase-app/pkg/types"

type GetAcquisitionRightsAggregateReq struct {
	UserId string `validate:"ulid"`
	types.MetaQuery

	AcademicManagerId string `query:"academic_manager_id" validate:"omitempty,ulid"`
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
	AcademicManagerId      string  `json:"academic_manager_id" db:"academic_manager_id"`
	AcademicManagerName    string  `json:"academic_manager_name" db:"academic_manager_name"`
	TotalAcquisitionRights float64 `json:"total_acquisition_rights" db:"total_acquisition_rights"`
}
