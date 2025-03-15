package seeds

import (
	"codebase-app/pkg/errmsg"
	"regexp"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/xuri/excelize/v2"
)

type excelSeed struct {
	db   *sqlx.DB
	file *excelize.File
}

func newExcelSeed(db *sqlx.DB) (*excelSeed, error) {
	f, err := excelize.OpenFile("db/seeds/excel/seeds.xlsx")
	if err != nil {
		log.Error().Err(err).Msg("failed to open excel file")
		return nil, err
	}

	return &excelSeed{db: db, file: f}, nil
}

func SeedExcel(db *sqlx.DB, sheetName string) error {
	excelSeeder, err := newExcelSeed(db)
	if err != nil {
		log.Error().Err(err).Msg("failed to create excel seeder")
		return err
	}

	tx, err := excelSeeder.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("failed to start transaction")
	}
	defer tx.Rollback()
	var errSeed error

	switch sheetName {
	case "roles":
		errSeed = excelSeeder.SeedRoles(tx)
		if errSeed != nil {
			return errSeed
		}
	case "permissions":
		errSeed = excelSeeder.SeedPermissions(tx)
		if errSeed != nil {
			return errSeed
		}
	case "programs":
		errSeed = excelSeeder.SeedPrograms(tx)
		if errSeed != nil {
			return errSeed
		}
	case "academic_managers":
		errSeed = excelSeeder.SeedAcademicManagers(tx)
		if errSeed != nil {
			return errSeed
		}
	case "lecturers":
		errSeed = excelSeeder.SeedLecturers(tx)
		if errSeed != nil {
			return errSeed
		}
	case "student_managers":
		errSeed = excelSeeder.SeedStudentManagers(tx)
		if errSeed != nil {
			return errSeed
		}
	case "marketers":
		errSeed = excelSeeder.SeedMarketers(tx)
		if errSeed != nil {
			return errSeed
		}
	case "students":
		errSeed = excelSeeder.SeedStudents(tx)
		if errSeed != nil {
			return errSeed
		}
	case "users":
		errSeed = excelSeeder.SeedUsers(tx)
		if errSeed != nil {
			return errSeed
		}
	}

	if err := excelSeeder.file.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save excel file")
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		return err
	}

	return nil
}

func (s *excelSeed) SeedPermissions(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("permissions")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1
	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id   = row[0]
			name = row[1]
			desc = row[2]
		)

		// if id is empty then add ULID to it, and when done it should be saved to db

		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("permissions", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO permissions (id, name, description) VALUES (?, ?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, name, desc)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert permission")
			return err
		}
	}

	// get all permissions from db that are not in the sheet
	type permission struct {
		Id   string  `db:"id"`
		Name string  `db:"name"`
		Desc *string `db:"description"`
	}
	var permissionsNotInSheet []permission

	query, args, err := sqlx.In("SELECT id, name FROM permissions WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for permissions not in sheet")
		return err
	}
	err = tx.Select(&permissionsNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get permissions not in sheet")
		return err
	}

	// append permissions not in sheet to the sheet
	for i, permission := range permissionsNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		cellC := "C" + rowNumber
		s.file.SetCellValue("permissions", cellA, permission.Id)
		s.file.SetCellValue("permissions", cellB, permission.Name)
		if permission.Desc != nil {
			s.file.SetCellValue("permissions", cellC, *permission.Desc)
		}
	}

	log.Info().Msg("permissions seeded successfully!")
	return nil
}

func (s *excelSeed) SeedPrograms(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("programs")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	// lastRow := len(rows) - 1
	re := regexp.MustCompile(`\D`)
	// insert into db
	for i, row := range rows {
		if len(row) < 5 {
			log.Error().Any("row", row).Msg("invalid row")
			return errmsg.NewCustomErrors(400).SetMessage("invalid row")
		}
		if i == 0 { // skip header
			continue
		}

		var (
			id                   = row[0]
			name                 = row[1]
			detailStr            = row[2]
			detail               *string
			priceStr             = row[3]
			price                float64
			daysStr              = row[4]
			commissionFee        float64
			commissionFeeStr     = row[5]
			lecturerFee          float64
			lecturerFeeStr       = row[6]
			feePerMeeting        float64
			feePerMeetingStr     = row[7]
			fullFee              float64
			fullFeeStr           = row[8]
			acquisitionRights    float64
			acquisitionRightsStr = row[9]
		)

		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("programs", cell, id)
		} else {
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}
		idsInSheet[i-1] = id

		if detailStr == "" {
			detail = nil
		} else {
			detail = &detailStr
		}

		if priceStr == "" {
			price = 0
		} else {
			priceFloat, err := strconv.ParseFloat(re.ReplaceAllString(priceStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse price")
				return err
			}
			price = priceFloat
		}

		if commissionFeeStr == "" {
			commissionFee = 0
		} else {
			commissionFeeFloat, err := strconv.ParseFloat(re.ReplaceAllString(commissionFeeStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse commission fee")
				return err
			}
			commissionFee = commissionFeeFloat
		}

		if lecturerFeeStr == "" {
			lecturerFee = 0
		} else {
			lecturerFeeFloat, err := strconv.ParseFloat(re.ReplaceAllString(lecturerFeeStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse lecturer fee")
				return err
			}
			lecturerFee = lecturerFeeFloat
		}

		if feePerMeetingStr == "" {
			feePerMeeting = 0
		} else {
			feePerMeetingFloat, err := strconv.ParseFloat(re.ReplaceAllString(feePerMeetingStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse fee per meeting")
				return err
			}
			feePerMeeting = feePerMeetingFloat
		}

		if fullFeeStr == "" {
			fullFee = 0
		} else {
			fullFeeFloat, err := strconv.ParseFloat(re.ReplaceAllString(fullFeeStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse full fee")
				return err
			}
			fullFee = fullFeeFloat
		}

		if acquisitionRightsStr == "" {
			acquisitionRights = 0
		} else {
			acquisitionRightsFloat, err := strconv.ParseFloat(re.ReplaceAllString(acquisitionRightsStr, ""), 64)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse acquisition rights")
				return err
			}
			acquisitionRights = acquisitionRightsFloat
		}

		// convert to []int
		days := make([]int, 0)
		if daysStr != "" { // example: "1|2|3"
			daysStrArr := strings.Split(daysStr, "|")
			for _, dayStr := range daysStrArr {
				day, err := strconv.Atoi(dayStr)
				if err != nil {
					log.Error().Err(err).Msg("failed to parse days")
					return err
				}
				days = append(days, day)
			}
		}

		query := `
			INSERT INTO programs (id, name, detail, price, days, commission_fee, lecturer_fee, price_per_meeting, full_fee, acquisition_rights)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				detail = EXCLUDED.detail,
				price = EXCLUDED.price,
				days = EXCLUDED.days,
				commission_fee = EXCLUDED.commission_fee,
				lecturer_fee = EXCLUDED.lecturer_fee,
				price_per_meeting = EXCLUDED.price_per_meeting,
				full_fee = EXCLUDED.full_fee,
				acquisition_rights = EXCLUDED.acquisition_rights
			`
		_, err := tx.Exec(s.db.Rebind(query), id, name, detail, price, pq.Array(days), commissionFee, lecturerFee, feePerMeeting, fullFee, acquisitionRights)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert program")
			return err
		}
	}

	log.Info().Msg("programs seeded successfully!")
	return nil
}

func (s *excelSeed) SeedAcademicManagers(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("academic_managers")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1

	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id   = row[0]
			name = row[1]
		)

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("academic_managers", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO academic_managers (id, name) VALUES (?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, name)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert academic manager")
			return err
		}

	}

	// get all academic managers from db that are not in the sheet
	type academicManager struct {
		Id   string `db:"id"`
		Name string `db:"name"`
	}

	var academicManagersNotInSheet []academicManager

	query, args, err := sqlx.In("SELECT id, name FROM academic_managers WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for academic managers not in sheet")
		return err
	}

	err = tx.Select(&academicManagersNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get academic managers not in sheet")
		return err
	}

	// append academic managers not in sheet to the sheet
	for i, academicManager := range academicManagersNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		s.file.SetCellValue("academic_managers", cellA, academicManager.Id)
		s.file.SetCellValue("academic_managers", cellB, academicManager.Name)
	}

	log.Info().Msg("academic managers seeded successfully!")
	return nil
}

func (s *excelSeed) SeedLecturers(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("lecturers")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1

	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id                = row[0]
			academicManagerId = row[1]
			name              = row[2]
			email             string
			phone             string
		)

		if isset(row, 3) {
			email = row[3]
		}
		if isset(row, 4) {
			phone = row[4]
		}

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("lecturers", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		if email == "" {
			email = gofakeit.Email()
		}
		if phone == "" {
			phone = gofakeit.Phone()
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO lecturers (id, academic_manager_id, name, phone, email) VALUES (?, ?, ?, ?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, academicManagerId, name, phone, email)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert lecturer")
			return err
		}
	}

	// get all lecturers from db that are not in the sheet
	type lecturer struct {
		Id                string  `db:"id"`
		AcademicManagerId string  `db:"academic_manager_id"`
		Name              string  `db:"name"`
		Phone             *string `db:"phone"`
		Email             *string `db:"email"`
	}

	var lecturersNotInSheet []lecturer

	query, args, err := sqlx.In("SELECT id, academic_manager_id, name, phone, email FROM lecturers WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for lecturers not in sheet")
		return err
	}

	err = tx.Select(&lecturersNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get lecturers not in sheet")
		return err
	}

	// append lecturers not in sheet to the sheet
	for i, lecturer := range lecturersNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		cellC := "C" + rowNumber
		cellD := "D" + rowNumber
		cellE := "E" + rowNumber
		s.file.SetCellValue("lecturers", cellA, lecturer.Id)
		s.file.SetCellValue("lecturers", cellB, lecturer.AcademicManagerId)
		s.file.SetCellValue("lecturers", cellC, lecturer.Name)
		if lecturer.Phone != nil {
			s.file.SetCellValue("lecturers", cellD, *lecturer.Phone)
		}
		if lecturer.Email != nil {
			s.file.SetCellValue("lecturers", cellE, *lecturer.Email)
		}
	}

	log.Info().Msg("lecturers seeded successfully!")
	return nil
}

func (s *excelSeed) SeedStudentManagers(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("student_managers")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1
	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id   = row[0]
			name = row[1]
		)

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("student_managers", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO student_managers (id, name) VALUES (?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, name)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert student manager")
			return err
		}
	}

	// get all student managers from db that are not in the sheet
	type studentManager struct {
		Id   string `db:"id"`
		Name string `db:"name"`
	}

	var studentManagersNotInSheet []studentManager

	query, args, err := sqlx.In("SELECT id, name FROM student_managers WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for student managers not in sheet")
		return err
	}

	err = tx.Select(&studentManagersNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get student managers not in sheet")
		return err
	}

	// append student managers not in sheet to the sheet
	for i, studentManager := range studentManagersNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		s.file.SetCellValue("student_managers", cellA, studentManager.Id)
		s.file.SetCellValue("student_managers", cellB, studentManager.Name)
	}

	log.Info().Msg("student managers seeded successfully!")
	return nil
}

func (s *excelSeed) SeedMarketers(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("marketers")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1
	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id               = row[0]
			studentManagerId = row[1]
			name             = row[2]
			email            string
			phone            string
		)

		if isset(row, 3) {
			email = row[3]
		}
		if isset(row, 4) {
			phone = row[4]
		}

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("marketers", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		if studentManagerId == "" {
			return errmsg.NewCustomErrors(400).SetMessage("student manager id is required")
		} else {
			// check student manager id is valid ULID
			if _, err := ulid.Parse(studentManagerId); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		if email == "" {
			email = gofakeit.Email()
		}
		if phone == "" {
			phone = gofakeit.Phone()
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO marketers (id, student_manager_id, name, phone, email) VALUES (?, ?, ?, ?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, studentManagerId, name, phone, email)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert marketer")
			return err
		}
	}

	// get all marketers from db that are not in the sheet
	type marketer struct {
		Id               string  `db:"id"`
		StudentManagerId string  `db:"student_manager_id"`
		Name             string  `db:"name"`
		Phone            *string `db:"phone"`
		Email            *string `db:"email"`
	}

	var marketersNotInSheet []marketer

	query, args, err := sqlx.In("SELECT id, student_manager_id, name, phone, email FROM marketers WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for marketers not in sheet")
		return err
	}

	err = tx.Select(&marketersNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get marketers not in sheet")
		return err
	}

	// append marketers not in sheet to the sheet
	for i, marketer := range marketersNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		cellC := "C" + rowNumber
		cellD := "D" + rowNumber
		cellE := "E" + rowNumber
		s.file.SetCellValue("marketers", cellA, marketer.Id)
		s.file.SetCellValue("marketers", cellB, marketer.StudentManagerId)
		s.file.SetCellValue("marketers", cellC, marketer.Name)
		if marketer.Phone != nil {
			s.file.SetCellValue("marketers", cellD, *marketer.Phone)
		}
		if marketer.Email != nil {
			s.file.SetCellValue("marketers", cellE, *marketer.Email)
		}
	}

	log.Info().Msg("marketers seeded successfully!")
	return nil
}

func (s *excelSeed) SeedStudents(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("students")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	idsInSheet := make([]string, len(rows)-1)
	lastRow := len(rows) - 1
	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id         = row[0]
			identifier = row[1]
			name       = row[2]
		)

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("students", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO students (id, identifier, name) VALUES (?, ?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, identifier, name)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert student")
			return err
		}
	}

	// get all students from db that are not in the sheet
	type student struct {
		Id         string `db:"id"`
		Identifier string `db:"identifier"`
		Name       string `db:"name"`
	}

	var studentsNotInSheet []student

	query, args, err := sqlx.In("SELECT id, identifier, name FROM students WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for students not in sheet")
		return err
	}

	err = tx.Select(&studentsNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get students not in sheet")
		return err
	}

	// append students not in sheet to the sheet
	for i, student := range studentsNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		cellC := "C" + rowNumber
		s.file.SetCellValue("students", cellA, student.Id)
		s.file.SetCellValue("students", cellB, student.Identifier)
		s.file.SetCellValue("students", cellC, student.Name)
	}

	log.Info().Msg("students seeded successfully!")
	return nil
}

func (s *excelSeed) SeedUsers(tx *sqlx.Tx) error {
	query := `
		INSERT INTO users (id, role_id, name, email, password)
		VALUES (
			?, (SELECT id FROM roles WHERE name = ?), ?, ?, ?
		)
		ON CONFLICT DO NOTHING
	`

	query = s.db.Rebind(query)
	password, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return err
	}

	// insert into db
	_, err = tx.Exec(query,
		ulid.Make().String(),
		"Super Admin",
		"super admin",
		"superadmin@kp.com",
		string(password),
	)

	if err != nil {
		log.Error().Err(err).Msg("failed to insert user")
		return err
	}

	log.Info().Msg("users seeded successfully!")
	return nil
}

func isset(arr []string, index int) bool {
	return (len(arr) > index)
}
