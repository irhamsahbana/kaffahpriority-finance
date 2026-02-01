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

func (s *payrollService) ImportPayrollItems(ctx context.Context, req *entity.ImportPayrollItemsReq) (*entity.ImportPayrollItemsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.ImportPayrollItems")
	defer span.End()

	const (
		sheetName = "Rekap Gaji"
		// colNo              = "A"
		colProgramMeetings = "F" // Jumlah Tatap Muka
		colFullWage        = "H" // Ujroh Full
		colWagePerMeeting  = "G" // Gaji per Tatap Muka / Hitungan
		// colInitialWage     = "I" // Ujroh Awal
		colIsMeetingFull = "J" // TF/F
		colForeignFee    = "K" // FL
		colNightFee      = "L" // NL
		colKeepWage      = "O" // Keep Gaji
		// colAcqRights       = "P" // Hak Akuisisi
		colID = "R" // Payroll Item ID
	)

	var errs = errmsg.NewCustomErrors(400)

	f, err := excelize.OpenReader(req.File)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("open file error")
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Ctx(ctx).Error().Err(cerr).Msg("close file error")
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("read rows error")
		return nil, fmt.Errorf("read rows: %w", err)
	}
	if len(rows) <= 1 {
		return &entity.ImportPayrollItemsResp{
			TotalProcessed: 0,
			TotalUpdated:   0,
		}, nil
	}

	req.Items = make([]entity.ImportedPayrollItems, 0, len(rows)-1)
	rawValue := excelize.Options{RawCellValue: true}

	// Start from row 2 (index 1)
	for i := 1; i < len(rows); i++ {
		rowIdx := i + 1

		// no, _ := f.GetCellValue(sheetName, cell(colNo, rowIdx))
		idStr, _ := f.GetCellValue(sheetName, cell(colID, rowIdx))
		meetingsStr, _ := f.GetCellValue(sheetName, cell(colProgramMeetings, rowIdx))
		wagePerMeetingStr, _ := f.GetCellValue(sheetName, cell(colWagePerMeeting, rowIdx), rawValue)
		fullWageStr, _ := f.GetCellValue(sheetName, cell(colFullWage, rowIdx), rawValue)
		// initialWageStr, _ := f.CalcCellValue(sheetName, cell(colInitialWage, rowIdx))
		isFullStr, _ := f.GetCellValue(sheetName, cell(colIsMeetingFull, rowIdx))
		flStr, _ := f.GetCellValue(sheetName, cell(colForeignFee, rowIdx), rawValue)
		nlStr, _ := f.GetCellValue(sheetName, cell(colNightFee, rowIdx), rawValue)
		keepWageStr, _ := f.GetCellValue(sheetName, cell(colKeepWage, rowIdx), rawValue)
		// acqStr, _ := f.GetCellValue(sheetName, cell(colAcqRights, rowIdx))

		idStr = strings.TrimSpace(idStr)
		meetingsStr = strings.TrimSpace(meetingsStr)
		fullWageStr = strings.TrimSpace(fullWageStr)
		fullWageStr = strings.ReplaceAll(fullWageStr, ",", "")
		wagePerMeetingStr = strings.TrimSpace(wagePerMeetingStr)
		wagePerMeetingStr = strings.ReplaceAll(wagePerMeetingStr, ",", "")
		// initialWageStr = strings.TrimSpace(initialWageStr)
		// initialWageStr = strings.ReplaceAll(initialWageStr, ",", "")
		isFullStr = strings.TrimSpace(isFullStr)
		flStr = strings.TrimSpace(flStr)
		nlStr = strings.TrimSpace(nlStr)
		keepWageStr = strings.TrimSpace(keepWageStr)
		keepWageStr = strings.ReplaceAll(keepWageStr, ",", "")
		// acqStr = strings.TrimSpace(acqStr)

		if idStr == "" {
			continue
		}

		item := entity.ImportedPayrollItems{}

		// Validate ID
		_, err := ulid.Parse(idStr)
		if err != nil {
			_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("ID tidak valid: %s", idStr))
		}
		item.ID = idStr

		// Is Meeting Full (TF/F)
		// Logic: "full" or "tidak full" as per Report module
		// But user said: "F" or "1" means Full.
		// Let's support both Report module standard AND User's specific request.
		// Report module: "full" -> true, "tidak full" -> false
		// User request: "F" or "1" -> true
		// I will normalize to lowercase.
		isFullLower := strings.ToLower(isFullStr)
		if isFullLower == "full" || isFullLower == "f" || isFullLower == "1" || isFullLower == "true" {
			item.IsMeetingFull = true
		} else {
			item.IsMeetingFull = false
		}

		fullWage := decimal.Zero
		if fullWageStr != "" {
			parsedFullWage, err := decimal.NewFromString(fullWageStr)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("Ujroh Full tidak valid: %s", fullWageStr))
			} else {
				fullWage = parsedFullWage
			}
		}
		if fullWage.LessThan(decimal.Zero) {
			_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), "Ujroh Full tidak boleh negatif")
		}
		item.FullWage = fullWage

		// Wage per meeting
		wagePerMeeting := decimal.Zero
		if wagePerMeetingStr != "" {
			parsedWagePerMeeting, err := decimal.NewFromString(wagePerMeetingStr)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("Gaji per Tatap Muka tidak valid: %s", wagePerMeetingStr))
			} else {
				wagePerMeeting = parsedWagePerMeeting
			}
		}

		// Program Meetings
		meetings, err := strconv.Atoi(meetingsStr)
		if err != nil {
			_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("Jumlah Tatap Muka tidak valid: %s", meetingsStr))
		}
		item.ProgramMeetings = meetings

		// Initial Wage
		if item.IsMeetingFull {
			item.InitialWage = fullWage
		} else {
			item.InitialWage = wagePerMeeting.Mul(decimal.NewFromInt(int64(meetings)))
		}

		// Foreign Learning Fee
		if flStr != "" {
			fl, err := decimal.NewFromString(flStr)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("FL tidak valid: %s", flStr))
			} else {
				item.ForeignLearningFee = &fl
			}
		} else {
			zero := decimal.Zero
			item.ForeignLearningFee = &zero
		}

		// Night Learning Fee
		if nlStr != "" {
			nl, err := decimal.NewFromString(nlStr)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("NL tidak valid: %s", nlStr))
			} else {
				item.NightLearningFee = &nl
			}
		} else {
			zero := decimal.Zero
			item.NightLearningFee = &zero
		}

		keepWage := decimal.Zero
		if keepWageStr != "" {
			parsedKeepWage, err := decimal.NewFromString(keepWageStr)
			if err != nil {
				_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), fmt.Sprintf("Keep Gaji tidak valid: %s", keepWageStr))
			} else {
				keepWage = parsedKeepWage
			}
		}
		if keepWage.LessThan(decimal.Zero) {
			_ = errs.Add(fmt.Sprintf("row_%d", rowIdx), "Keep Gaji tidak boleh negatif")
		}
		item.Wage = keepWage
		if keepWage.Equal(decimal.Zero) {
			item.AcquisitionRights = nil
		}

		req.Items = append(req.Items, item)
	}

	if errs.HasErrors() {
		return nil, errs
	}

	rowsAffected, err := s.repo.ImportPayrollItems(ctx, req)
	if err != nil {
		return nil, err
	}

	return &entity.ImportPayrollItemsResp{
		TotalProcessed: len(req.Items),
		TotalUpdated:   int(rowsAffected),
	}, nil
}

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}
