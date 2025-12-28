package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *reportRepo) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.UseHRfeeForLecturer")
	defer span.End()
	var (
		fnName = "repo::UseHRfeeForLecturer"
	)

	monthlyPool, err := r.GetSameMonthLecturerRegistrations(ctx, req.RegistrationID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get monthly pool", fnName)
		return err
	}

	// Calculate total available budget in the pool
	var totalPoolBudget decimal.Decimal
	for _, item := range monthlyPool {
		totalPoolBudget = totalPoolBudget.Add(item.MentorDetailFee)
		// Console log to help user see what's happening in 'task dev'
		fmt.Printf("[DEBUG] %s - Pool Item: ID=%s, Category=%s, Budget=%s, ExistingUsed=%v\n",
			fnName, item.ID, item.Category, item.MentorDetailFee.String(), item.MentorDetailFeeUsed)
	}

	requestedAmt := decimal.Zero
	if req.UsedAmount != nil {
		requestedAmt = *req.UsedAmount
	}

	fmt.Printf("[DEBUG] %s - RegistrationID: %s, RequestedTotal: %s, TotalPoolBudget: %s, PoolSize: %d\n",
		fnName, req.RegistrationID, requestedAmt.String(), totalPoolBudget.String(), len(monthlyPool))

	// Validation: If total requested used amount > total combined budget, error.
	if req.UsedAmount != nil && requestedAmt.GreaterThan(totalPoolBudget) {
		registrationResp, err := r.GetRegistration(ctx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.RegistrationID,
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get registration info", fnName)
			return err
		}

		lecturerName := "Mentor belum dipilih"
		if registrationResp.LecturerName != nil {
			lecturerName = *registrationResp.LecturerName
		}

		errMsg := fmt.Sprintf(
			"Jumlah yang diminta (%s) melebihi total budget tersedia (%s) untuk kelompok pendaftaran ini (Santri: %s, Program: %s, Mentor: %s).",
			requestedAmt.String(),
			totalPoolBudget.String(),
			registrationResp.StudentName,
			registrationResp.ProgramName,
			lecturerName,
		)

		fmt.Printf("[ERROR] %s - Validation failed: %s\n", fnName, errMsg)
		log.Ctx(ctx).Error().Any("req", req).Any("pool_budget", totalPoolBudget).Msgf("%s - validation failed", fnName)
		return errmsg.
			NewCustomErrors(http.StatusUnprocessableEntity).
			Add("used_amount", errMsg).
			SetMessage(errMsg)
	}

	// If req.UsedAmount is nil, clear all records in the specific month's pool
	if req.UsedAmount == nil {
		Tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer Tx.Rollback()

		ids := make([]string, 0, len(monthlyPool))
		for _, item := range monthlyPool {
			ids = append(ids, item.ID)
		}

		queryClear := `UPDATE program_registrations SET mentor_detail_fee_used = NULL, notes_for_fund_distributions = NULL WHERE id IN (?)`
		query, args, err := sqlx.In(queryClear, ids)
		if err != nil {
			return err
		}

		_, err = Tx.ExecContext(ctx, Tx.Rebind(query), args...)
		if err != nil {
			return err
		}

		return Tx.Commit()
	}

	// Start distribution logic
	Tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer Tx.Rollback()

	remainingToAllocate := requestedAmt

	// Sort: General first, then target ID, then stability
	sort.Slice(monthlyPool, func(i, j int) bool {
		if monthlyPool[i].Category == "general" && monthlyPool[j].Category != "general" {
			return true
		}
		if monthlyPool[i].Category != "general" && monthlyPool[j].Category == "general" {
			return false
		}
		if monthlyPool[i].ID == req.RegistrationID {
			return true
		}
		if monthlyPool[j].ID == req.RegistrationID {
			return false
		}
		return monthlyPool[i].ID < monthlyPool[j].ID
	})

	for _, item := range monthlyPool {
		alloc := decimal.Min(remainingToAllocate, item.MentorDetailFee)
		remainingToAllocate = remainingToAllocate.Sub(alloc)

		if item.ID == req.RegistrationID {
			queryUpdate := `UPDATE program_registrations SET mentor_detail_fee_used = LEAST(?, mentor_detail_fee), notes_for_fund_distributions = ? WHERE id = ? AND deleted_at IS NULL`
			_, err = Tx.ExecContext(ctx, Tx.Rebind(queryUpdate), alloc, req.Notes, item.ID)
		} else {
			queryOther := `UPDATE program_registrations SET mentor_detail_fee_used = LEAST(?, mentor_detail_fee) WHERE id = ? AND deleted_at IS NULL`
			_, err = Tx.ExecContext(ctx, Tx.Rebind(queryOther), alloc, item.ID)
		}
		if err != nil {
			return err
		}
	}

	return Tx.Commit()
}

func (r *reportRepo) GetSameMonthLecturerRegistrations(ctx context.Context, registrationID string) ([]entity.RelatedRegistration, error) {
	items := make([]entity.RelatedRegistration, 0)
	query := `
		WITH target AS (
			SELECT
				pr.id,
				COALESCE(pr.lecturer_id, parent.lecturer_id) as lecturer_id,
				COALESCE(pr.allocated_at, parent.allocated_at) as allocated_at,
				COALESCE(pr.parent_id, pr.id) as cluster_root_id
			FROM
				program_registrations pr
			LEFT JOIN
				program_registrations parent ON pr.parent_id = parent.id
			WHERE
				pr.deleted_at IS NULL
				AND (pr.parent_id IS NULL OR parent.deleted_at IS NULL)
				AND pr.id = ?
		),
		pool_candidates AS (
			SELECT
				pr.id,
				pr.category,
				pr.mentor_detail_fee,
				pr.mentor_detail_fee_used,
				COALESCE(pr.lecturer_id, parent.lecturer_id) as lecturer_id,
				COALESCE(pr.allocated_at, parent.allocated_at) as allocated_at,
				COALESCE(pr.parent_id, pr.id) as cluster_root_id
			FROM
				program_registrations pr
			LEFT JOIN
				program_registrations parent ON pr.parent_id = parent.id
			WHERE
				pr.deleted_at IS NULL
				AND (pr.parent_id IS NULL OR parent.deleted_at IS NULL)
				AND pr.is_paid = TRUE
		)
		SELECT
			pc.id,
			pc.category,
			pc.mentor_detail_fee,
			pc.mentor_detail_fee_used
		FROM
			pool_candidates pc
		CROSS JOIN
			target t
		WHERE
			pc.id = t.id -- Always include the record being edited
			OR
			(
				t.lecturer_id IS NOT NULL 
				AND pc.lecturer_id = t.lecturer_id 
				AND pc.cluster_root_id = t.cluster_root_id -- Limit to the same billing cluster
				AND TO_CHAR(pc.allocated_at AT TIME ZONE 'Asia/Makassar', 'YYYY-MM') = 
				    TO_CHAR(t.allocated_at AT TIME ZONE 'Asia/Makassar', 'YYYY-MM')
			)
	`
	err := r.db.SelectContext(ctx, &items, r.db.Rebind(query), registrationID)
	if err == nil {
		fmt.Printf("[DEBUG] repo::GetSameMonthLecturerRegistrations - fetched %d items for reg_id %s\n", len(items), registrationID)
	}
	return items, err
}

func (r *reportRepo) BulkUseHRfeeForLecturer(ctx context.Context, req *entity.BulkUseHRfeeForLecturerReq) error {
	var (
		fnName  = "repo::BulkUseHRfeeForLecturer"
		errBulk = []error{}
	)

	for _, item := range req.Data {
		err := r.UseHRfeeForLecturer(ctx, &item)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Any("item", item).Msgf("%s - failed to use hr fee for lecturer", fnName)
			errBulk = append(errBulk, err)
		}

	}

	if len(errBulk) > 0 {
		var errorMessages string
		for _, err := range errBulk {
			errorMessages += err.Error() + "\n"
		}

		return errmsg.NewCustomErrors(http.StatusUnprocessableEntity).
			SetMessage(errorMessages)
	}

	return nil
}
