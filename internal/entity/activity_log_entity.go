package entity

import (
	"codebase-app/pkg/types"
	"time"
)

type ActivityLogType string

const (
	ActivityLogTypeCreate ActivityLogType = "create"
	ActivityLogTypeUpdate ActivityLogType = "update"
	ActivityLogTypeDelete ActivityLogType = "delete"
	// ActivityLogTypeComment ActivityLogType = "comment"
)

type ActivityLogEntity string

const (
	ActivityLogEntityPrograms                     ActivityLogEntity = "programs"
	ActivityLogEntityProgramRegistrationTemplates ActivityLogEntity = "program_registration_templates"
	ActivityLogEntityProgramRegistrations         ActivityLogEntity = "program_registrations"
)

type ActivityLog struct {
	ID           string               `json:"id" db:"id"`
	EntityID     string               `json:"entity_id" db:"entity_id"`
	EntityName   ActivityLogEntity    `json:"entity_name" db:"entity_name"`
	Type         ActivityLogType      `json:"type" db:"type"`
	AuthorName   string               `json:"author_name" db:"author_name"`
	AuthorRole   string               `json:"author_role" db:"author_role"`
	Changes      []ActivityLogChanges `json:"changes" db:"changes"`
	Before       any                  `json:"before" db:"before"`
	After        any                  `json:"after" db:"after"`
	Message      string               `json:"message" db:"message"`
	ErrorMessage string               `json:"error_message" db:"error_message"`
	CreatedAt    time.Time            `json:"created_at" db:"created_at"`
	CreatedBy    string               `json:"created_by" db:"created_by"`
	UpdatedAt    time.Time            `json:"updated_at" db:"updated_at"`
	UpdatedBy    string               `json:"updated_by" db:"updated_by"`
	DeletedAt    *time.Time           `json:"deleted_at" db:"deleted_at"`
	DeletedBy    *string              `json:"deleted_by" db:"deleted_by"`
}

type ActivityLogChanges struct {
	Field  string `json:"field" db:"field"`
	Before string `json:"before" db:"before"`
	After  string `json:"after" db:"after"`
	Note   string `json:"note" db:"note"`
}

type GetActivityLogsReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	EntityID   string `query:"entity_id" validate:"required"`
	EntityName string `query:"entity_name" validate:"required"`
	SortBy     string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at paid_at student_name"`
	SortType   string `query:"sort_type" validate:"omitempty,oneof=asc desc"`
	types.MetaQuery
}
