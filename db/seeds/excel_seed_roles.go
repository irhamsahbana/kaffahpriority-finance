package seeds

import (
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

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
