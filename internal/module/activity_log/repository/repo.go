package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	ports "codebase-app/internal/ports/module/activity_log"
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (r *activityLogRepo) GetActivityLogs(ctx context.Context, req *entity.GetActivityLogsReq) (*entity.GetActivityLogsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetActivityLogs")
	defer span.End()

	type dao struct {
		TotalData int       `db:"total_data"`
		ID        string    `db:"id"`
		Logs      []byte    `db:"logs"`
		CreatedAt time.Time `db:"created_at"`
	}
	var (
		data = make([]dao, 0, req.Paginate)
		resp = new(entity.GetActivityLogsResp)
		args = make([]any, 0, 3)
	)
	resp.Items = make([]entity.ActivityLogList, 0, req.Paginate)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			id, logs, created_at
			FROM activity_logs
		WHERE
			1 = 1
	`

	if len(req.EntityIDs) > 0 {
		query += " AND entity_id = ANY(?)"
		args = append(args, pq.Array(req.EntityIDs))
	}

	if req.EntityName != "" {
		query += " AND entity_name = ?"
		args = append(args, req.EntityName)
	}

	if len(req.ActivityTypes) > 0 {
		query += " AND logs->>'type' = ANY(?)"
		args = append(args, pq.Array(req.ActivityTypes))
	}

	query += `
		ORDER BY ` + req.SortBy + ` ` + req.SortType + `
	`

	query += `
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, req.Paginate*(req.Page-1))

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("error select activity log")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		var logs entity.ActivityLog
		if err := json.Unmarshal(item.Logs, &logs); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("item", item).Msg("error unmarshal activity log")
			continue
		}
		resp.Items = append(resp.Items, entity.ActivityLogList{
			ActivityLog:  logs,
			LogID:        item.ID,
			LogCreatedAt: item.CreatedAt,
		})
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
