package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func (s *payrollService) ExportPayrollItemsPeriodically(ctx context.Context, req *entity.ExportPayrollItemsPeriodicallyReq) (*entity.ExportPayrollItemsPeriodicallyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.ExportPayrollItemsPeriodically")
	defer span.End()

	// 1. Get Payroll Run for the period
	run, err := s.repo.GetPayrollRunByPeriod(ctx, req.Period)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll run by period")
		return nil, err
	}

	// 2. Get Payroll Items
	items, err := s.repo.GetPayrollItemsForExport(ctx, run.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll items for export")
		return nil, err
	}

	// 3. Fetch Additional Students
	var itemIDs []string
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}

	additionalStudentsMap, err := s.repo.GetAdditionalStudents(ctx, itemIDs)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get additional students")
		return nil, err
	}

	// Enrich items with additional students
	for i := range items {
		if students, ok := additionalStudentsMap[items[i].ID]; ok {
			for _, s := range students {
				items[i].StudentName += ", " + s.Name
			}
		}
	}

	// 4. Group Data
	type groupedLecturer struct {
		LecturerName string
		Items        []entity.PayrollItem
	}

	type groupedAcademicManager struct {
		AcademicManagerName string
		Lecturers           []groupedLecturer
	}

	var groupedData []groupedAcademicManager

	if len(items) > 0 {
		var currentAM *groupedAcademicManager
		var currentLecturer *groupedLecturer

		for _, item := range items {
			// Check if AM changed
			if currentAM == nil || currentAM.AcademicManagerName != item.AcademicManagerName {
				// Push previous AM if exists
				if currentAM != nil {
					if currentLecturer != nil {
						currentAM.Lecturers = append(currentAM.Lecturers, *currentLecturer)
					}
					groupedData = append(groupedData, *currentAM)
				}
				currentAM = &groupedAcademicManager{
					AcademicManagerName: item.AcademicManagerName,
					Lecturers:           []groupedLecturer{},
				}
				currentLecturer = nil // Reset lecturer when AM changes
			}

			// Check if Lecturer changed
			if currentLecturer == nil || currentLecturer.LecturerName != item.LecturerName {
				if currentLecturer != nil {
					currentAM.Lecturers = append(currentAM.Lecturers, *currentLecturer)
				}
				currentLecturer = &groupedLecturer{
					LecturerName: item.LecturerName,
					Items:        []entity.PayrollItem{},
				}
			}

			// Add item
			currentLecturer.Items = append(currentLecturer.Items, item)
		}

		// Push the last ones
		if currentAM != nil {
			if currentLecturer != nil {
				currentAM.Lecturers = append(currentAM.Lecturers, *currentLecturer)
			}
			groupedData = append(groupedData, *currentAM)
		}
	}

	// 5. Generate Excel
	f := excelize.NewFile()
	var sheetName = "Sheet1"

	// Styles
	HeaderStyle, _ := newHeaderStyle(f, "#3BFFF5", true)
	HeaderStyleRight, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Border: defaultBorderStyle(),
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#3BFFF5"},
			Pattern: 1,
		},
		NumFmt: 3,
	})
	academicManagerNameStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  24,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	editableStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFF2CC"},
			Pattern: 1,
		},
	})

	// Headering
	f.SetCellValue(sheetName, "A1", "NO")
	f.SetCellValue(sheetName, "B1", "NAMA")
	f.SetCellValue(sheetName, "C1", "NO")
	f.SetCellValue(sheetName, "D1", "NAMA")
	f.SetCellValue(sheetName, "E1", "PROGRAM")
	f.SetCellValue(sheetName, "F1", "JUMLAH")
	f.SetCellValue(sheetName, "G1", "HITUNGAN")
	f.SetCellValue(sheetName, "H1", "UJROH FULL")
	f.SetCellValue(sheetName, "I1", "UJROH AWAL")
	f.SetCellValue(sheetName, "J1", "TF/F")
	f.SetCellValue(sheetName, "K1", "FL")
	f.SetCellValue(sheetName, "L1", "NL")
	f.SetCellValue(sheetName, "M1", "UJROH REAL")
	f.SetCellValue(sheetName, "N1", "KETERANGAN")
	f.SetCellValue(sheetName, "O1", "KEEP GAJI")
	f.SetCellValue(sheetName, "P1", "HAK AKUISISI")
	f.SetCellValue(sheetName, "Q1", "PENASEHAT AKADEMIK")

	f.SetCellStyle(sheetName, "A1", "Q1", HeaderStyle)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "G", "I", 20)
	f.SetColWidth(sheetName, "M", "M", 20)
	f.SetColWidth(sheetName, "N", "N", 20)
	f.SetColWidth(sheetName, "O", "O", 20)
	f.SetColWidth(sheetName, "P", "P", 20)
	f.SetColWidth(sheetName, "Q", "Q", 20)

	_ = f.SetPanes(sheetName, &excelize.Panes{
		Freeze: true,
		XSplit: 5,
		YSplit: 1,
	})

	lastRow := 1

	for _, am := range groupedData {
		lastRow++
		f.SetCellValue(sheetName, fmt.Sprintf("A%v", lastRow), am.AcademicManagerName)
		f.MergeCell(sheetName, fmt.Sprintf("A%v", lastRow), fmt.Sprintf("Q%v", lastRow+1))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%v", lastRow), fmt.Sprintf("Q%v", lastRow+1), academicManagerNameStyle)
		lastRow++

		for lecturerIndex, lecturer := range am.Lecturers {
			lastRow++
			f.SetCellValue(sheetName, fmt.Sprintf("A%v", lastRow), lecturerIndex+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow), lecturer.LecturerName)

			var totalRealFee float64

			for itemIndex, item := range lecturer.Items {
				f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow), itemIndex+1)
				f.SetCellValue(sheetName, fmt.Sprintf("D%v", lastRow), item.StudentName)
				f.SetCellValue(sheetName, fmt.Sprintf("E%v", lastRow), item.ProgramName)
				f.SetCellValue(sheetName, fmt.Sprintf("Q%v", lastRow), item.MarketerName)

				f.SetCellValue(sheetName, fmt.Sprintf("F%v", lastRow), item.ProgramMeetings)
				f.SetCellValue(sheetName, fmt.Sprintf("G%v", lastRow), item.WagePerMeeting.InexactFloat64())
				f.SetCellValue(sheetName, fmt.Sprintf("H%v", lastRow), item.FullWage.InexactFloat64())
				f.SetCellValue(sheetName, fmt.Sprintf("I%v", lastRow), item.InitialWage.InexactFloat64())

				isFullStr := "TF"
				if item.IsMeetingFull {
					isFullStr = "F"
				}
				f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow), isFullStr)

				if !item.ForeignLearningFee.IsZero() {
					f.SetCellValue(sheetName, fmt.Sprintf("K%v", lastRow), item.ForeignLearningFee.InexactFloat64())
				}
				if !item.NightLearningFee.IsZero() {
					f.SetCellValue(sheetName, fmt.Sprintf("L%v", lastRow), item.NightLearningFee.InexactFloat64())
				}

				wage := item.Wage.InexactFloat64()
				f.SetCellValue(sheetName, fmt.Sprintf("M%v", lastRow), wage)
				totalRealFee += wage

				// Keep gaji
				_ = f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow), wage)

				// Hak akuisisi
				_ = f.SetCellFormula(sheetName, fmt.Sprintf("P%v", lastRow), fmt.Sprintf(`IF(O%v=0,"",%d)`, lastRow, item.AcquisitionRights))

				// ID for re-import (Hidden or explicit column R)
				f.SetCellValue(sheetName, fmt.Sprintf("R%v", lastRow), item.ID)

				// Apply styles to editable columns
				f.SetCellStyle(sheetName, fmt.Sprintf("F%v", lastRow), fmt.Sprintf("F%v", lastRow), editableStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("I%v", lastRow), fmt.Sprintf("I%v", lastRow), editableStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("J%v", lastRow), fmt.Sprintf("J%v", lastRow), editableStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("K%v", lastRow), fmt.Sprintf("K%v", lastRow), editableStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("L%v", lastRow), fmt.Sprintf("L%v", lastRow), editableStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("O%v", lastRow), fmt.Sprintf("O%v", lastRow), editableStyle)

				if itemIndex+1 != len(lecturer.Items) {
					lastRow++
				}

				if itemIndex == len(lecturer.Items)-1 {
					lastRow++
					f.SetCellValue(sheetName, fmt.Sprintf("L%v", lastRow), "TOTAL")
					f.SetCellValue(sheetName, fmt.Sprintf("M%v", lastRow), totalRealFee)
					f.SetCellStyle(sheetName, fmt.Sprintf("L%v", lastRow), fmt.Sprintf("L%v", lastRow), HeaderStyle)
					f.SetCellStyle(sheetName, fmt.Sprintf("M%v", lastRow), fmt.Sprintf("M%v", lastRow), HeaderStyleRight)
					// lastRow++
				}
			}
		}
	}

	tmstmp := time.Now().Unix()
	filename := fmt.Sprintf("payroll-items-%s-%v.xlsx", req.Period, tmstmp)
	filepath := "./" + filename

	if err := f.SaveAs(filepath); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error saving file")
		return nil, err
	}

	return &entity.ExportPayrollItemsPeriodicallyResp{
		FilePath: filepath,
		FileName: filename,
	}, nil
}

func defaultBorderStyle() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "#000000", Style: 1},
		{Type: "top", Color: "#000000", Style: 1},
		{Type: "bottom", Color: "#000000", Style: 1},
		{Type: "right", Color: "#000000", Style: 1},
	}
}

func newHeaderStyle(f *excelize.File, fillColor string, withNumFmt bool) (int, error) {
	style := &excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: defaultBorderStyle(),
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{fillColor},
			Pattern: 1,
		},
	}
	if withNumFmt {
		style.NumFmt = 3
	}
	return f.NewStyle(style)
}
