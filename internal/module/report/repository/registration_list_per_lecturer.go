package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"sort"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegistrationListPerLecturer
	}

	type argCombine struct {
		LecturerId *string
		StudentId  string
		ProgramId  string
	}

	var (
		data         = make([]dao, 0)
		dataPerMonth = make([]entity.RegistrationListPerLecturerPerMonth, 0)
		args         = make([]any, 0, 3)
		resp         = new(entity.GetRegistrationListPerLecturerResp)
		argsCombine  = make([]argCombine, 0)
	)
	resp.Items = make([]entity.RegistrationListPerLecturer, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			pr.program_id,
			pr.lecturer_id,
			pr.student_id,
			l.name AS lecturer_name,
			s.name AS student_name,
			p.name AS program_name
		FROM
			program_registrations pr
		JOIN
			programs p ON pr.program_id = p.id
		JOIN
			students s ON pr.student_id = s.id
		LEFT JOIN
			lecturers l ON pr.lecturer_id = l.id
		WHERE 1 = 1
	`
	if req.Q != "" {
		query += ` AND (l.name ILIKE ? OR s.name ILIKE ?)`
		args = append(args, "%"+req.Q+"%", "%"+req.Q+"%")
	}

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	if req.StudentId != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentId)
	}

	query += `
		GROUP BY
			pr.program_id,
			pr.lecturer_id,
			pr.student_id,
			l.name,
			s.name,
			p.name
		ORDER BY
			pr.lecturer_id ASC,
			pr.student_id ASC,
			p.name ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrationListPerLecturer - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Items = append(resp.Items, item.RegistrationListPerLecturer)
		resp.Items[len(resp.Items)-1].Registrations = make([]entity.RegistrationListPerLecturerPerMonth, 0)
		resp.Items[len(resp.Items)-1].Year = req.Year
		resp.Meta.TotalData = item.TotalData

		argsCombine = append(argsCombine, argCombine{
			LecturerId: item.LecturerId,
			StudentId:  item.StudentId,
			ProgramId:  item.ProgramId,
		})
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	// query for registration per month
	if len(resp.Items) == 0 {
		return resp, nil
	}

	args = make([]any, 0, 3)

	query = `
		WITH months AS (
			SELECT 1 AS month_num, 'Januari' AS month_name UNION ALL
			SELECT 2, 'Februari' UNION ALL
			SELECT 3, 'Maret' UNION ALL
			SELECT 4, 'April' UNION ALL
			SELECT 5, 'Mei' UNION ALL
			SELECT 6, 'Juni' UNION ALL
			SELECT 7, 'Juli' UNION ALL
			SELECT 8, 'Agustus' UNION ALL
			SELECT 9, 'September' UNION ALL
			SELECT 10, 'Oktober' UNION ALL
			SELECT 11, 'November' UNION ALL
			SELECT 12, 'Desember'
		)
		SELECT
			m.month_name AS month,
			m.month_num,
			pr.id AS registration_id,
			pr.mentor_detail_fee AS hr_fee_for_lecturer,
			pr.mentor_detail_fee_used AS used_amount,
			CASE
				WHEN pr.mentor_detail_fee_used IS NOT NULL THEN TRUE
				ELSE NULL
			END AS is_used,
			pr.notes_for_fund_distributions AS notes,

			pr.program_id,
			pr.lecturer_id,
			pr.student_id,
			pr.foreign_learning_fee,
			pr.night_learning_fee
		FROM
			program_registrations pr
		JOIN
			months m
			ON EXTRACT(MONTH FROM (pr.started_at AT TIME ZONE ?)) = m.month_num
		WHERE
			pr.deleted_at IS NULL
			AND EXTRACT(YEAR FROM (pr.started_at AT TIME ZONE ?)) = ?
		`

	args = append(args, req.Tz, req.Tz, req.Year)

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	if req.StudentId != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentId)
	}

	if len(argsCombine) > 0 {
		query += ` AND (`

		for i, item := range argsCombine {
			if i > 0 {
				query += ` OR `
			}

			if item.LecturerId != nil {
				query += ` (pr.lecturer_id = ? AND pr.student_id = ? AND pr.program_id = ?) `
				args = append(args, *item.LecturerId, item.StudentId, item.ProgramId)
			} else {
				query += ` (pr.student_id = ? AND pr.program_id = ? AND pr.lecturer_id IS NULL) `
				args = append(args, item.StudentId, item.ProgramId)
			}
		}

		query += ` )`
	}

	query += `
		ORDER BY
			pr.lecturer_id ASC,
			pr.student_id ASC,
			m.month_num ASC
	`

	err = r.db.SelectContext(ctx, &dataPerMonth, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrationListPerLecturer - failed to fetch data")
		return nil, err
	}

	var allMonths = []struct {
		Num  int
		Name string
	}{
		{1, "Januari"}, {2, "Februari"}, {3, "Maret"}, {4, "April"},
		{5, "Mei"}, {6, "Juni"}, {7, "Juli"}, {8, "Agustus"},
		{9, "September"}, {10, "Oktober"}, {11, "November"}, {12, "Desember"},
	}

	for i := range resp.Items {
		// Buat map untuk menyimpan data per bulan
		monthMap := make(map[int]entity.RegistrationListPerLecturerPerMonth)

		// Isi map dengan data hasil query
		for _, item := range dataPerMonth {
			var lecturerId1, lecturerId2 string
			if resp.Items[i].LecturerId != nil {
				lecturerId1 = *resp.Items[i].LecturerId
			}
			if item.LecturerId != nil {
				lecturerId2 = *item.LecturerId
			}

			if (resp.Items[i].LecturerId == nil && item.LecturerId == nil) ||
				(resp.Items[i].LecturerId != nil && item.LecturerId != nil && lecturerId1 == lecturerId2) &&
					resp.Items[i].StudentId == item.StudentId &&
					resp.Items[i].ProgramId == item.ProgramId {
				monthMap[item.MonthNum] = item
			}
		}

		// Pastikan setiap bulan dari Januari-Desember ada di hasil akhir
		for _, m := range allMonths {
			if _, exists := monthMap[m.Num]; !exists {
				// Tambahkan data default untuk bulan yang tidak ada
				resp.Items[i].Registrations = append(resp.Items[i].Registrations, entity.RegistrationListPerLecturerPerMonth{
					Month:      m.Name,
					MonthNum:   m.Num,
					IsUsed:     nil, // Nilai default jika tidak ada data
					UsedAmount: nil, // Nilai default jika tidak ada data
					Notes:      nil, // Nilai default jika tidak ada data
					ProgramId:  resp.Items[i].ProgramId,
					LecturerId: resp.Items[i].LecturerId,
					StudentId:  resp.Items[i].StudentId,
				})
			} else {
				// Jika bulan sudah ada dalam data, tambahkan ke Registrations
				resp.Items[i].Registrations = append(resp.Items[i].Registrations, monthMap[m.Num])

				// tambahkan keterangan FL dan NL
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].FL != nil {
					resp.Items[i].IsFL = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].NL != nil {
					resp.Items[i].IsNL = true
				}
			}
		}

		// Urutkan berdasarkan nomor bulan setelah pengisian
		sort.Slice(resp.Items[i].Registrations, func(a, b int) bool {
			return resp.Items[i].Registrations[a].MonthNum < resp.Items[i].Registrations[b].MonthNum
		})
	}

	return resp, nil
}
