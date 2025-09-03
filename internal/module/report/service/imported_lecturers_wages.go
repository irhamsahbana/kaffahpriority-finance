package service

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func (s *reportService) ImportLecturersWages(
	ctx context.Context, // ctx, context untuk lifecycle request
	req *entity.ImportLecturersWagesReq, // req, memuat path file Excel dan opsi simpan
) error {
	const (
		fnName       = "service::ImportLecturersWages"
		sheetName    = "Sheet1"
		colNo        = "A"
		colJumlah    = "F" // editable
		colUjrohAwal = "I" // editable
		colTFF       = "J" // editable
		colFL        = "K" // editable
		colNL        = "L" // editable
		colNotes     = "N" // editable
		colID        = "R" // registrationId, key update
	)

	var errs = errmsg.NewCustomErrors(400)

	f, err := excelize.OpenReader(req.File)
	if err != nil {
		log.Error().Err(err).Msgf("%s - open file error", fnName)
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Error().Err(cerr).Msgf("%s - close file error", fnName)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Error().Err(err).Msgf("%s - read rows error", fnName)
		return fmt.Errorf("read rows: %w", err)
	}
	if len(rows) <= 1 {
		// tidak ada data, hanya header atau kosong
		return nil
	}

	req.Registrations = make([]entity.ImportedLecturersWages, 0, len(rows)-1)
	rawValue := excelize.Options{RawCellValue: true}

	// Mulai dari baris 2, asumsikan baris 1 header.
	for i := 1; i < len(rows); i++ {
		rowIdx := i + 1 // indeks Excel satuan, dimulai dari 1

		no, _ := f.GetCellValue(sheetName, cell(colNo, rowIdx))
		registrationId, _ := f.GetCellValue(sheetName, cell(colID, rowIdx))
		jumlahStr, _ := f.GetCellValue(sheetName, cell(colJumlah, rowIdx))            // F
		awalStr, _ := f.GetCellValue(sheetName, cell(colUjrohAwal, rowIdx), rawValue) // I
		tfFlag, _ := f.GetCellValue(sheetName, cell(colTFF, rowIdx))                  // J
		flStr, _ := f.GetCellValue(sheetName, cell(colFL, rowIdx), rawValue)          // K
		nlStr, _ := f.GetCellValue(sheetName, cell(colNL, rowIdx), rawValue)          // L
		notesStr, _ := f.GetCellValue(sheetName, cell(colNotes, rowIdx))              // N

		registrationId = strings.TrimSpace(registrationId)
		jumlahStr = strings.TrimSpace(jumlahStr)
		awalStr = strings.TrimSpace(awalStr)
		tfFlag = strings.TrimSpace(tfFlag)
		flStr = strings.TrimSpace(flStr)
		nlStr = strings.TrimSpace(nlStr)
		notesStr = strings.TrimSpace(notesStr)

		// skip baris kosong
		if registrationId == "" {
			continue
		}

		data := entity.ImportedLecturersWages{}

		// field RegistrationId
		_, err := ulid.Parse(registrationId)
		if err != nil {
			errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("registration_id bukan ULID yang valid: %s", registrationId))
		}
		data.RegistrationID = registrationId

		// field ProgramMeetings
		programMeetings, err := strconv.Atoi(jumlahStr)
		if err != nil {
			errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("jumlah bukan angka yang valid: %s", jumlahStr))
		}
		data.ProgramMeetings = programMeetings

		// field InitialFee
		initialFee, err := decimal.NewFromString(awalStr)
		if err != nil {
			errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("ujroh awal bukan angka yang valid: %s", awalStr))
		}
		if initialFee.LessThan(decimal.Zero) {
			errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("ujroh awal tidak boleh negatif: %s", awalStr))
		}
		data.InitialFee = initialFee

		// field IsFullFee
		if tfFlag != "full" && tfFlag != "tidak full" {
			errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), "TF/F hanya boleh berisi 'full' atau 'tidak full'")
		}
		switch tfFlag {
		case "full":
			data.IsFullFee = true
		case "tidak full":
			data.IsFullFee = false
		}

		// field FL
		if flStr != "" {
			fl, err := decimal.NewFromString(flStr)
			if err != nil {
				errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("FL bukan angka yang valid: %s", flStr))
			}
			if fl.LessThan(decimal.Zero) {
				errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("FL tidak boleh negatif: %s", flStr))
			}
			data.FL = &fl
		}

		// field NL
		if nlStr != "" {
			nl, err := decimal.NewFromString(nlStr)
			if err != nil {
				errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("NL bukan angka yang valid: %s", nlStr))
			}
			if nl.LessThan(decimal.Zero) {
				errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("NL tidak boleh negatif: %s", nlStr))
			}
			data.NL = &nl
		}

		// field notes
		if notesStr != "" {
			if len(notesStr) > 255 {
				errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), "notes tidak boleh lebih dari 255 karakter")
			}
			data.Notes = &notesStr
		}

		req.Registrations = append(req.Registrations, data)
	}

	if errs.HasErrors() {
		log.Warn().Err(errs).Any("errors", errs).Msgf("%s - invalid request", fnName)
		return errs
	}

	if err := s.repo.ImportLecturersWages(ctx, req); err != nil {
		return err
	}

	return nil
}

// util

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}
