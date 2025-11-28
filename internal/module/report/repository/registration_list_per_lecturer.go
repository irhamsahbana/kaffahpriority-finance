package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	fnName := "repo::GetRegistrationsPerLecturer"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegistrationListPerLecturer
	}

	type argCombine struct {
		LecturerID *string
		StudentID  string
		ProgramID  string
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
			COUNT(*) OVER () AS total_data,
			prt.program_id,
			prt.lecturer_id,
			prt.student_id,
			l.name AS lecturer_name,
			s.name AS student_name,
			p.name AS program_name,
			l.academic_manager_id,
			am.name AS academic_manager_name
		FROM
			program_registration_templates prt
		JOIN
			programs p ON prt.program_id = p.id
		JOIN
			students s ON prt.student_id = s.id
		LEFT JOIN
			lecturers l ON prt.lecturer_id = l.id
		LEFT JOIN
			academic_managers am ON l.academic_manager_id = am.id
		LEFT JOIN
			program_registrations pr ON prt.program_id = pr.program_id 
			AND prt.lecturer_id = pr.lecturer_id 
			AND prt.student_id = pr.student_id
			AND pr.deleted_at IS NULL
		WHERE prt.deleted_at IS NULL
	`

	if req.Q != "" {
		query += ` AND (l.name ILIKE ? OR s.name ILIKE ?)`
		args = append(args, "%"+req.Q+"%", "%"+req.Q+"%")
	}

	if req.LecturerID != "" {
		query += ` AND prt.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += ` AND prt.student_id = ?`
		args = append(args, req.StudentID)
	}

	if req.AcademicManagerID != "" {
		query += ` AND l.academic_manager_id = ?`
		args = append(args, req.AcademicManagerID)
	}

	query += `
		GROUP BY
			prt.program_id,
			prt.lecturer_id,
			prt.student_id,
			l.name,
			s.name,
			p.name,
			l.academic_manager_id,
			am.name,
			l.id
		ORDER BY
			l.academic_manager_id ASC,
			l.id ASC,
			MIN(prt.created_at) ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	for _, item := range data {
		resp.Items = append(resp.Items, item.RegistrationListPerLecturer)
		resp.Items[len(resp.Items)-1].Registrations = make([]entity.RegistrationListPerLecturerPerMonth, 0)
		resp.Items[len(resp.Items)-1].Year = req.Year
		resp.Meta.TotalData = item.TotalData

		argsCombine = append(argsCombine, argCombine{
			LecturerID: item.LecturerID,
			StudentID:  item.StudentID,
			ProgramID:  item.ProgramID,
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
			pr.night_learning_fee,
			pr.is_itp,
			CASE
				WHEN pr.program_meetings > 0 THEN TRUE
				ELSE FALSE
			END AS is_started
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			months m
			ON EXTRACT(MONTH FROM (pr.allocated_at AT TIME ZONE ?)) = m.month_num
		WHERE
			pr.deleted_at IS NULL
			AND pr.is_paid = TRUE
			AND EXTRACT(YEAR FROM (pr.allocated_at AT TIME ZONE ?)) = ?
		`

	args = append(args, req.Tz, req.Tz, req.Year)

	if req.AcademicManagerID != "" {
		query += ` AND l.academic_manager_id = ?`
		args = append(args, req.AcademicManagerID)
	}

	if req.LecturerID != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentID)
	}

	if len(argsCombine) > 0 {
		query += ` AND (`

		for i, item := range argsCombine {
			if i > 0 {
				query += ` OR `
			}

			if item.LecturerID != nil {
				query += ` (pr.lecturer_id = ? AND pr.student_id = ? AND pr.program_id = ?) `
				args = append(args, *item.LecturerID, item.StudentID, item.ProgramID)
			} else {
				query += ` (pr.student_id = ? AND pr.program_id = ? AND pr.lecturer_id IS NULL) `
				args = append(args, item.StudentID, item.ProgramID)
			}
		}

		query += ` )`
	}

	query += `
		ORDER BY
			l.academic_manager_id ASC,
			pr.lecturer_id ASC,
			m.month_num ASC
	`

	err = r.db.SelectContext(ctx, &dataPerMonth, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data per month", fnName)
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

	// Buat Map untuk menyimpan registrasi paling akhir di tiap grup
	// Key: index dari resp.Items
	// Value: registrasi paling akhir
	lastRegistrations := make(map[int]*string)
	lastRegistrationWithKeyId := make(map[string]int)
	registrationIds := make([]string, 0)

	for i := range resp.Items {
		// Buat map untuk menyimpan data per bulan
		monthMap := make(map[int]entity.RegistrationListPerLecturerPerMonth)

		// Isi map dengan data hasil query
		for _, item := range dataPerMonth {
			var lecturerId1, lecturerId2 string
			if resp.Items[i].LecturerID != nil {
				lecturerId1 = *resp.Items[i].LecturerID
			}
			if item.LecturerID != nil {
				lecturerId2 = *item.LecturerID
			}

			if (resp.Items[i].LecturerID == nil && item.LecturerID == nil) ||
				(resp.Items[i].LecturerID != nil && item.LecturerID != nil && lecturerId1 == lecturerId2) &&
					resp.Items[i].StudentID == item.StudentID &&
					resp.Items[i].ProgramID == item.ProgramID {
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
					ProgramID:  resp.Items[i].ProgramID,
					LecturerID: resp.Items[i].LecturerID,
					StudentID:  resp.Items[i].StudentID,
				})
			} else {
				// Jika bulan sudah ada dalam data, tambahkan ke Registrations
				resp.Items[i].Registrations = append(resp.Items[i].Registrations, monthMap[m.Num])

				// tambahkan keterangan FL, NL, ITP dan IsStarted pada bulan terakhir
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].FL != nil {
					resp.Items[i].IsFL = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].NL != nil {
					resp.Items[i].IsNL = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].IsITP {
					resp.Items[i].IsITP = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].IsStarted {
					resp.Items[i].IsStarted = true
				}

				// Simpan data registrasi terakhir
				lastRegistrations[i] = resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].RegistrationID
				registrationIds = append(registrationIds, *lastRegistrations[i])

				// Simpan index registrasi terakhir
				lastRegistrationWithKeyId[*lastRegistrations[i]] = i
			}
		}

		// Urutkan berdasarkan nomor bulan setelah pengisian
		sort.Slice(resp.Items[i].Registrations, func(a, b int) bool {
			return resp.Items[i].Registrations[a].MonthNum < resp.Items[i].Registrations[b].MonthNum
		})
	}

	// Jika tidak ada data registrasi, kembalikan langsung kembalikan
	// response tanpa melakukan query tambahan
	if len(registrationIds) == 0 {
		return resp, nil
	}

	// query for additional students
	query = `
		SELECT
			pr.id,
			STRING_AGG(pras.name, ', ') AS additional_students
		FROM
			program_registrations pr
		JOIN
			pr_additional_students pras ON pr.id = pras.pr_id
		WHERE
			pr.id IN (?)
		GROUP BY
			pr.id
	`

	type additionalStudents struct {
		RegistrationId     string `db:"id"`
		AdditionalStudents string `db:"additional_students"`
	}

	var additionalStudentsData = make([]additionalStudents, 0)

	query, args, err = sqlx.In(query, registrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - error preparing query for additional students", fnName)
		return nil, err
	}

	err = r.db.SelectContext(ctx, &additionalStudentsData, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
		return nil, err
	}

	// Tambahkan data additional students ke resp.Items field student_name
	for _, item := range additionalStudentsData {
		if idx, exists := lastRegistrationWithKeyId[item.RegistrationId]; exists {
			resp.Items[idx].StudentName += ", " + item.AdditionalStudents
		}
	}

	return resp, nil
}

func (r *reportRepo) GetRegistrationsPerLecturerV2(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	fnName := "repo::GetRegistrationsPerLecturerV2"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegistrationListPerLecturer
	}

	type argCombine struct {
		LecturerID *string
		StudentID  string
		ProgramID  string
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
			COUNT(*) OVER () AS total_data,
			prt.program_id,
			prt.lecturer_id,
			prt.student_id,
			l.name AS lecturer_name,
			s.name AS student_name,
			p.name AS program_name,
			l.academic_manager_id,
			am.name AS academic_manager_name
		FROM
			program_registration_templates prt
		JOIN
			programs p ON prt.program_id = p.id
		JOIN
			students s ON prt.student_id = s.id
		LEFT JOIN
			lecturers l ON prt.lecturer_id = l.id
		LEFT JOIN
			academic_managers am ON l.academic_manager_id = am.id
		LEFT JOIN
			program_registrations pr ON prt.program_id = pr.program_id 
			AND prt.lecturer_id = pr.lecturer_id 
			AND prt.student_id = pr.student_id
			AND pr.deleted_at IS NULL
		WHERE prt.deleted_at IS NULL
	`

	if req.Q != "" {
		query += ` AND (l.name ILIKE ? OR s.name ILIKE ?)`
		args = append(args, "%"+req.Q+"%", "%"+req.Q+"%")
	}

	if req.LecturerID != "" {
		query += ` AND prt.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += ` AND prt.student_id = ?`
		args = append(args, req.StudentID)
	}

	if req.AcademicManagerID != "" {
		query += ` AND l.academic_manager_id = ?`
		args = append(args, req.AcademicManagerID)
	}

	query += `
		GROUP BY
			prt.program_id,
			prt.lecturer_id,
			prt.student_id,
			l.name,
			s.name,
			p.name,
			l.academic_manager_id,
			am.name,
			l.id
		ORDER BY
			l.academic_manager_id ASC,
			l.id ASC,
			MIN(prt.created_at) ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	for _, item := range data {
		resp.Items = append(resp.Items, item.RegistrationListPerLecturer)
		resp.Items[len(resp.Items)-1].Registrations = make([]entity.RegistrationListPerLecturerPerMonth, 0)
		resp.Items[len(resp.Items)-1].Year = req.Year
		resp.Meta.TotalData = item.TotalData

		argsCombine = append(argsCombine, argCombine{
			LecturerID: item.LecturerID,
			StudentID:  item.StudentID,
			ProgramID:  item.ProgramID,
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
		),
		pr_base AS (
			SELECT pr.*, COALESCE(pr.parent_id, pr.id) AS root_id
			FROM program_registrations pr
		),
		root_info AS (
			SELECT
				p.id AS root_id,
				EXTRACT(MONTH FROM (p.allocated_at AT TIME ZONE ?)) AS month_num,
				EXTRACT(YEAR FROM (p.allocated_at AT TIME ZONE ?)) AS year_num,
				l.academic_manager_id,
				p.lecturer_id AS root_lecturer_id,
				p.program_id AS root_program_id,
				p.student_id AS root_student_id
			FROM program_registrations p
			LEFT JOIN lecturers l ON p.lecturer_id = l.id
		),
		combined AS (
			SELECT b.*, ri.month_num, ri.year_num, ri.academic_manager_id,
				ri.root_lecturer_id, ri.root_program_id, ri.root_student_id,
				(
					SELECT p.is_paid FROM program_registrations p WHERE p.id = ri.root_id
				) AS root_is_paid,
				(
					SELECT p.deleted_at FROM program_registrations p WHERE p.id = ri.root_id
				) AS root_deleted_at
			FROM pr_base b
			JOIN root_info ri ON b.root_id = ri.root_id
		)
		SELECT
			m.month_name AS month,
			m.month_num,
			c.root_id AS registration_id,
			SUM(c.mentor_detail_fee) AS hr_fee_for_lecturer,
			SUM(c.mentor_detail_fee_used) AS used_amount,
			CASE
				WHEN SUM(CASE WHEN c.mentor_detail_fee_used IS NOT NULL AND c.mentor_detail_fee_used > 0 THEN 1 ELSE 0 END) > 0 THEN TRUE
				ELSE NULL
			END AS is_used,
			MAX(c.notes_for_fund_distributions) AS notes,

			c.root_program_id AS program_id,
			c.root_lecturer_id AS lecturer_id,
			c.root_student_id AS student_id,
			MAX(c.foreign_learning_fee) AS foreign_learning_fee,
			MAX(c.night_learning_fee) AS night_learning_fee,
			BOOL_OR(c.is_itp) AS is_itp,
			CASE
				WHEN MAX(c.program_meetings) > 0 THEN TRUE
				ELSE FALSE
			END AS is_started
		FROM
			combined c
		JOIN
			months m ON c.month_num = m.month_num
		WHERE
			c.deleted_at IS NULL
			AND c.root_deleted_at IS NULL
			AND c.root_is_paid = TRUE
			AND c.year_num = ?
		`

	args = append(args, req.Tz, req.Tz, req.Year)

	if req.AcademicManagerID != "" {
		query += ` AND c.academic_manager_id = ?`
		args = append(args, req.AcademicManagerID)
	}

	if req.LecturerID != "" {
		query += ` AND c.root_lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += ` AND c.root_student_id = ?`
		args = append(args, req.StudentID)
	}

	if len(argsCombine) > 0 {
		query += ` AND (`

		for i, item := range argsCombine {
			if i > 0 {
				query += ` OR `
			}

			if item.LecturerID != nil {
				query += ` (c.root_lecturer_id = ? AND c.root_student_id = ? AND c.root_program_id = ?) `
				args = append(args, *item.LecturerID, item.StudentID, item.ProgramID)
			} else {
				query += ` (c.root_student_id = ? AND c.root_program_id = ? AND c.root_lecturer_id IS NULL) `
				args = append(args, item.StudentID, item.ProgramID)
			}
		}

		query += ` )`
	}

	query += `
		GROUP BY
			m.month_name,
			m.month_num,
			c.root_id,
			c.root_program_id,
			c.root_lecturer_id,
			c.root_student_id,
			c.academic_manager_id
		ORDER BY
			c.academic_manager_id ASC,
			c.root_lecturer_id ASC,
			m.month_num ASC
	`

	err = r.db.SelectContext(ctx, &dataPerMonth, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data per month", fnName)
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

	// Buat Map untuk menyimpan registrasi paling akhir di tiap grup
	// Key: index dari resp.Items
	// Value: registrasi paling akhir
	lastRegistrations := make(map[int]*string)
	lastRegistrationWithKeyId := make(map[string]int)
	registrationIds := make([]string, 0)

	for i := range resp.Items {
		// Buat map untuk menyimpan data per bulan
		monthMap := make(map[int]entity.RegistrationListPerLecturerPerMonth)

		// Isi map dengan data hasil query
		for _, item := range dataPerMonth {
			var lecturerId1, lecturerId2 string
			if resp.Items[i].LecturerID != nil {
				lecturerId1 = *resp.Items[i].LecturerID
			}
			if item.LecturerID != nil {
				lecturerId2 = *item.LecturerID
			}

			if (resp.Items[i].LecturerID == nil && item.LecturerID == nil) ||
				(resp.Items[i].LecturerID != nil && item.LecturerID != nil && lecturerId1 == lecturerId2) &&
					resp.Items[i].StudentID == item.StudentID &&
					resp.Items[i].ProgramID == item.ProgramID {
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
					ProgramID:  resp.Items[i].ProgramID,
					LecturerID: resp.Items[i].LecturerID,
					StudentID:  resp.Items[i].StudentID,
				})
			} else {
				// Jika bulan sudah ada dalam data, tambahkan ke Registrations
				resp.Items[i].Registrations = append(resp.Items[i].Registrations, monthMap[m.Num])

				// tambahkan keterangan FL, NL, ITP dan IsStarted pada bulan terakhir
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].FL != nil {
					resp.Items[i].IsFL = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].NL != nil {
					resp.Items[i].IsNL = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].IsITP {
					resp.Items[i].IsITP = true
				}
				if resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].IsStarted {
					resp.Items[i].IsStarted = true
				}

				// Simpan data registrasi terakhir
				lastRegistrations[i] = resp.Items[i].Registrations[len(resp.Items[i].Registrations)-1].RegistrationID
				registrationIds = append(registrationIds, *lastRegistrations[i])

				// Simpan index registrasi terakhir
				lastRegistrationWithKeyId[*lastRegistrations[i]] = i
			}
		}

		// Urutkan berdasarkan nomor bulan setelah pengisian
		sort.Slice(resp.Items[i].Registrations, func(a, b int) bool {
			return resp.Items[i].Registrations[a].MonthNum < resp.Items[i].Registrations[b].MonthNum
		})
	}

	// Jika tidak ada data registrasi, kembalikan langsung kembalikan
	// response tanpa melakukan query tambahan
	if len(registrationIds) == 0 {
		return resp, nil
	}

	// query for additional students
	query = `
		SELECT
			pr.id,
			STRING_AGG(pras.name, ', ') AS additional_students
		FROM
			program_registrations pr
		JOIN
			pr_additional_students pras ON pr.id = pras.pr_id
		WHERE
			pr.id IN (?)
		GROUP BY
			pr.id
	`

	type additionalStudents struct {
		RegistrationId     string `db:"id"`
		AdditionalStudents string `db:"additional_students"`
	}

	var additionalStudentsData = make([]additionalStudents, 0)

	query, args, err = sqlx.In(query, registrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - error preparing query for additional students", fnName)
		return nil, err
	}

	err = r.db.SelectContext(ctx, &additionalStudentsData, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
		return nil, err
	}

	// Tambahkan data additional students ke resp.Items field student_name
	for _, item := range additionalStudentsData {
		if idx, exists := lastRegistrationWithKeyId[item.RegistrationId]; exists {
			resp.Items[idx].StudentName += ", " + item.AdditionalStudents
		}
	}

	return resp, nil
}
