package seeds

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
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
		// case "users":
		// 	errSeed = excelSeeder.SeedUsers(tx)
		// 	if errSeed != nil {
		// 		return errSeed
		// 	}
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

func (s *excelSeed) SeedRoles(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("roles")
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
			s.file.SetCellValue("roles", cell, id)
		} else {
			// check id is valid ULID
			if _, err := ulid.Parse(id); err != nil {
				log.Error().Err(err).Msg("invalid ULID")
				return err
			}
		}

		idsInSheet[i-1] = id

		query := "INSERT INTO roles (id, name) VALUES (?, ?) ON CONFLICT DO NOTHING"
		_, err := tx.Exec(s.db.Rebind(query), id, name)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert role")
			return err
		}
	}

	// get all roles from db that are not in the sheet
	type role struct {
		Id   string `db:"id"`
		Name string `db:"name"`
	}
	var rolesNotInSheet []role

	query, args, err := sqlx.In("SELECT id, name FROM roles WHERE id NOT IN (?)", idsInSheet)
	if err != nil {
		log.Error().Err(err).Msg("failed to create query for roles not in sheet")
		return err
	}
	err = tx.Select(&rolesNotInSheet, s.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Msg("failed to get roles not in sheet")
		return err
	}

	// append roles not in sheet to the sheet
	for i, role := range rolesNotInSheet {
		rowNumber := strconv.Itoa(lastRow + i + 2)
		cellA := "A" + rowNumber
		cellB := "B" + rowNumber
		s.file.SetCellValue("roles", cellA, role.Id)
		s.file.SetCellValue("roles", cellB, role.Name)
	}

	log.Info().Msg("roles seeded successfully!")

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

func (s *excelSeed) SeedUsers(tx *sqlx.Tx) error {
	rows, err := s.file.GetRows("users")
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows from excel")
		return err
	}

	// insert into db
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}

		var (
			id          = row[0]
			name        = row[1]
			branchName  = row[2]
			sectionName = row[3]
			RoleName    = row[4]
			email       = row[5]
			password    = row[6]
		)

		// manipulate data
		switch strings.ToUpper(RoleName) {
		case "SERVICE ADVISOR":
			RoleName = "service_advisor"
		case "MRA":
			RoleName = "technician"
		case "ADMIN":
			RoleName = "admin"
		}

		// bcrypt password
		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("failed to hash password")
			return err
		}

		// if id is empty then add ULID to it, and when done it should be saved to db
		// and file
		if id == "" {
			id = ulid.Make().String()
			rowNumber := strconv.Itoa(i + 1)
			cell := "A" + rowNumber
			s.file.SetCellValue("users", cell, id)
		}

		// make sure email is in lowercase
		email = strings.ToLower(email)

		// check id is valid ULID
		if _, err := ulid.Parse(id); err != nil {
			log.Error().Err(err).Msg("invalid ULID")
			return err
		}

		switch strings.ToUpper(RoleName) {
		case "ADMIN":
			query := `
			INSERT INTO users (
				id, role_id, name, email, password
			) VALUES (
				?,
				(SELECT id FROM roles WHERE name = ?),
				?,
				?,
				?
			) ON CONFLICT (id) DO UPDATE SET
				role_id = (SELECT id FROM roles WHERE name = ?),
				name = ?,
				email = ?,
				password = ?
		`

			_, err = tx.Exec(s.db.Rebind(query),
				id, RoleName, name, email, string(passwordHashed),
				RoleName, name, email, string(passwordHashed),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to insert user")
				return err
			}
		default:
			//  on conflict update
			query := `
			INSERT INTO users (
				id, name, branch_id, section_id, role_id, email, password
			) VALUES (
				?,
				?,
				(SELECT id FROM branches WHERE UPPER(name) = UPPER(?)),
				(SELECT id FROM potencies WHERE UPPER(name) = UPPER(?)),
				(SELECT id FROM roles WHERE name = ?),
				?,
				?
			) ON CONFLICT (id) DO UPDATE SET
				name = ?,
				branch_id = (SELECT id FROM branches WHERE UPPER(name) = UPPER(?)),
				section_id = (SELECT id FROM potencies WHERE UPPER(name) = UPPER(?)),
				role_id = (SELECT id FROM roles WHERE name = ?),
				email = ?,
				password = ?
		`

			_, err = tx.Exec(s.db.Rebind(query),
				id, name, branchName, sectionName, RoleName, email, string(passwordHashed),
				name, branchName, sectionName, RoleName, email, string(passwordHashed),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to insert user")
				return err
			}
		}
	}

	log.Info().Msg("users seeded successfully!")

	return nil
}
