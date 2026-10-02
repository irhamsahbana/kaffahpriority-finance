package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
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
	ctx, span := tracing.StartSpan(ctx, "service.ImportLecturersWages")
	defer span.End()

	const (
		fnName    = "service::ImportLecturersWages"
		sheetName = "Sheet1"
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

	// Map header name -> column letter, so the parser survives column
	// reordering between export versions (old files have no PC-LN columns).
	headerToCol := map[string]string{}
	for i, header := range rows[0] {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}
		colLetter, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		headerToCol[header] = colLetter
	}
	lookupCol := func(name string) (string, bool) {
		col, ok := headerToCol[name]
		return col, ok
	}

	// required columns (exist in both old and new export layouts)
	colJumlah, okJumlah := lookupCol("Jumlah")
	colUjrohAwal, okAwal := lookupCol("Ujroh Awal")
	colTFF, okTFF := lookupCol("TF/F")
	colID, okID := lookupCol("ID")
	if !okJumlah || !okAwal || !okTFF || !okID {
		log.Error().Any("headers", headerToCol).Msgf("%s - required columns not found", fnName)
		return errmsg.NewCustomErrors(400).SetMessage("format file tidak valid: kolom Jumlah, Ujroh Awal, TF/F, atau ID tidak ditemukan")
	}

	// optional columns (PC-LN only exist in the new export layout)
	colFL, okFL := lookupCol("FL")
	colNL, okNL := lookupCol("NL")
	colPC, okPC := lookupCol("PC")
	colMT, okMT := lookupCol("MT")
	colCL, okCL := lookupCol("CL")
	colMS, okMS := lookupCol("MS")
	colSC, okSC := lookupCol("SC")
	colLN, okLN := lookupCol("LN")
	colNotes, okNotes := lookupCol("Keterangan")

	req.Registrations = make([]entity.ImportedLecturersWages, 0, len(rows)-1)
	rawValue := excelize.Options{RawCellValue: true}

	// Mulai dari baris 2, asumsikan baris 1 header.
	for i := 1; i < len(rows); i++ {
		rowIdx := i + 1 // indeks Excel satuan, dimulai dari 1

		no := fmt.Sprintf("%v", rowIdx)
		registrationId, _ := f.GetCellValue(sheetName, cell(colID, rowIdx))
		jumlahStr, _ := f.GetCellValue(sheetName, cell(colJumlah, rowIdx))   // F
		awalStr, _ := f.CalcCellValue(sheetName, cell(colUjrohAwal, rowIdx)) // I
		tfFlag, _ := f.GetCellValue(sheetName, cell(colTFF, rowIdx))         // J

		registrationId = strings.TrimSpace(registrationId)
		jumlahStr = strings.TrimSpace(jumlahStr)
		awalStr = strings.TrimSpace(awalStr)
		awalStr = strings.ReplaceAll(awalStr, ",", "")
		tfFlag = strings.TrimSpace(tfFlag)

		// skip baris kosong
		if registrationId == "" {
			continue
		}

		data := entity.ImportedLecturersWages{}

		// field RegistrationId
		_, err := ulid.Parse(registrationId)
		if err != nil {
			_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("registration_id bukan ULID yang valid: %s", registrationId))
		}
		data.RegistrationID = registrationId

		// field ProgramMeetings
		programMeetings, err := strconv.Atoi(jumlahStr)
		if err != nil {
			_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("jumlah bukan angka yang valid: %s", jumlahStr))
		}
		data.ProgramMeetings = programMeetings

		// field InitialFee
		initialFee, err := decimal.NewFromString(awalStr)
		if err != nil {
			_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("ujroh awal bukan angka yang valid: %s", awalStr))
		}
		if initialFee.LessThan(decimal.Zero) {
			_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("ujroh awal tidak boleh negatif: %s", awalStr))
		}
		data.InitialFee = initialFee

		// field IsFullFee
		if tfFlag != "full" && tfFlag != "tidak full" {
			_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), "TF/F hanya boleh berisi 'full' atau 'tidak full'")
		}
		switch tfFlag {
		case "full":
			data.IsFullFee = true
		case "tidak full":
			data.IsFullFee = false
		}

		// parseOptionalFee membaca kolom fee opsional (kosong = tidak diupdate)
		parseOptionalFee := func(col, label string) *decimal.Decimal {
			if col == "" {
				return nil
			}
			str, _ := f.GetCellValue(sheetName, cell(col, rowIdx), rawValue)
			str = strings.TrimSpace(str)
			if str == "" {
				return nil
			}
			val, err := decimal.NewFromString(str)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("%s bukan angka yang valid: %s", label, str))
				return nil
			}
			if val.LessThan(decimal.Zero) {
				_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), fmt.Sprintf("%s tidak boleh negatif: %s", label, str))
				return nil
			}
			return &val
		}

		if okFL {
			data.FL = parseOptionalFee(colFL, "FL")
		}
		if okNL {
			data.NL = parseOptionalFee(colNL, "NL")
		}
		if okPC {
			data.PC = parseOptionalFee(colPC, "PC")
		}
		if okMT {
			data.MT = parseOptionalFee(colMT, "MT")
		}
		if okCL {
			data.CL = parseOptionalFee(colCL, "CL")
		}
		if okMS {
			data.MS = parseOptionalFee(colMS, "MS")
		}
		if okSC {
			data.SC = parseOptionalFee(colSC, "SC")
		}
		if okLN {
			data.LN = parseOptionalFee(colLN, "LN")
		}

		// field notes
		if okNotes {
			notesStr, _ := f.GetCellValue(sheetName, cell(colNotes, rowIdx)) // T
			notesStr = strings.TrimSpace(notesStr)
			if notesStr != "" {
				if len(notesStr) > 255 {
					_ = errs.Add(fmt.Sprintf("file.%s.%s", sheetName, no), "notes tidak boleh lebih dari 255 karakter")
				}
				data.Notes = &notesStr
			}
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
