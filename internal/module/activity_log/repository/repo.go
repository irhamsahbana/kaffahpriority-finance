package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	ports "codebase-app/internal/ports/module/activity_log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var _ ports.ActivityLogRepository = &activityLogRepo{}

type activityLogRepo struct {
	db *sqlx.DB
}

func NewActivityLogRepository() *activityLogRepo {
	return &activityLogRepo{
		db: adapter.Adapters.Postgres,
	}
}

func (r *activityLogRepo) CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error {
	ctx, span := tracing.StartSpan(ctx, "repo.CreateActivityLog")
	defer span.End()

	fnName := "repo::CreateActivityLog"

	queryInsert := `
		INSERT INTO activity_logs (
			id, entity_name, entity_id, logs) VALUES (
			$1, $2, $3, $4
		)
	`

	jsonLogs, err := json.Marshal(req)
	if err != nil {
		log.Error().Err(err).Str("fn", fnName).Msg("error marshal activity log")
		return err
	}

	_, err = r.db.ExecContext(ctx, queryInsert,
		req.ID,
		req.EntityName,
		req.EntityID,
		jsonLogs,
	)
	if err != nil {
		log.Error().Err(err).Str("fn", fnName).Msg("error insert activity log")
		return err
	}

	return nil
}

func (r *activityLogRepo) GetActivityLogs(ctx context.Context, req *entity.GetActivityLogsReq) ([]entity.ActivityLog, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetActivityLog")
	defer span.End()

	fnName := "repo::GetActivityLog"

	query := `
		SELECT id, logs FROM activity_logs
		WHERE entity_name = $1 AND entity_id = $2
	`

	if req.SortBy != "" {
		orderBy := "created_at"
		if req.SortBy == "created_at" {
			orderBy = "logs->>'created_at'"
		}

		orderDir := "DESC"
		if req.SortType == "asc" {
			orderDir = "ASC"
		}

		query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
	}

	var resp []entity.ActivityLog
	rows, err := r.db.QueryxContext(ctx, query, req.EntityName, req.EntityID)
	if err != nil && err != sql.ErrNoRows {
		log.Error().Err(err).Str("fn", fnName).Msg("error get activity log")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rawLog []byte
		var logs entity.ActivityLog
		if err := rows.Scan(&logs.ID, &rawLog); err != nil {
			log.Error().Err(err).Str("fn", fnName).Msg("error scan activity log")
			return nil, err
		}

		if err := json.Unmarshal(rawLog, &logs); err != nil {
			log.Error().Err(err).Str("fn", fnName).Msg("error unmarshal activity log")
			return nil, err
		}
		resp = append(resp, logs)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("fn", fnName).Msg("error iterate activity log rows")
		return nil, err
	}

	if err != nil {
		log.Warn().Err(err).Str("fn", fnName).Msg("activity log not found")
	}

	return resp, nil
}
