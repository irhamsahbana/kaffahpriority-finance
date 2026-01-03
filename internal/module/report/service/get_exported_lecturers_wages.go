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

func (s *reportService) GetExportedLecturersWages(ctx context.Context, req *entity.GetExportedLecturersWagesReq) (*entity.GetExportedLecturersWagesResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetExportedLecturersWages")
	defer span.End()
	resp, err := s.repo.GetExportedLecturersWages(ctx, req)
	if err != nil {
		return nil, err
	}

	var (
		fnName    = "service::GetExportedLecturersWages"
		sheetName = "Sheet1"
	)

	f := excelize.NewFile()

	headerStyle, _ := newHeaderStyle(f, "#FFFF00", false)
	headerEditableStyle, _ := newHeaderStyle(f, "#00aef9ff", false)
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3,
	})
	currencyBoldStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3,
		Font:   &excelize.Font{Bold: true},
	})

	f.NewSheet(sheetName)

	// column width
	f.SetCellValue(sheetName, "A1", "No.")
	f.SetCellValue(sheetName, "B1", "MA")
	f.SetCellValue(sheetName, "C1", "Mentor (Pengajar)")
	f.SetCellValue(sheetName, "D1", "Nama Santri")
	f.SetCellValue(sheetName, "E1", "Program")
	f.SetCellValue(sheetName, "F1", "Jumlah") // editable
	f.SetCellValue(sheetName, "G1", "Hitungan")
	f.SetCellValue(sheetName, "H1", "Ujroh Full")
	f.SetCellValue(sheetName, "I1", "Ujroh Awal") // editable
	f.SetCellValue(sheetName, "J1", "TF/F")       // editable
	f.SetCellValue(sheetName, "K1", "FL")         // editable
	f.SetCellValue(sheetName, "L1", "NL")         // editable
	f.SetCellValue(sheetName, "M1", "Ujroh Real")
	f.SetCellValue(sheetName, "N1", "Keterangan") // editable
	f.SetCellValue(sheetName, "O1", "Keep Gaji")
	f.SetCellValue(sheetName, "P1", "Angka")
	f.SetCellValue(sheetName, "Q1", "MPA")
	f.SetCellValue(sheetName, "R1", "ID")
	f.SetColWidth(sheetName, "B", "Q", 20)

	currencyCols := []string{"G", "H", "I", "K", "L", "M", "O"}
	for _, col := range currencyCols {
		f.SetColStyle(sheetName, col, currencyStyle)
	}

	f.SetCellStyle(sheetName, "A1", "Q1", headerStyle)
	editableHeaders := []string{"F", "I", "J", "K", "L"}
	for _, col := range editableHeaders {
		f.SetCellStyle(sheetName, col+"1", col+"1", headerEditableStyle)
	}

	f.SetPanes("Sheet1", &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      5,    // Freeze kolom A-E (5 kolom pertama)
		YSplit:      1,    // Freeze baris pertama
		TopLeftCell: "F2", // Cell yang terlihat di pojok kiri atas setelah freeze
		ActivePane:  "bottomRight",
	})

	// data
	var prevLecturer string
	var prevAcademicManager string
	var amTotalRealFee float64
	var row = 2
	var seq = 1
	for index, item := range resp.Items {
		// handle lecturer saat ini (aman dari nil)
		curLecturer := ""
		if item.LecturerName != nil {
			curLecturer = *item.LecturerName
		}

		curAcademicManager := ""
		if item.AcademicManagerName != nil {
			curAcademicManager = *item.AcademicManagerName
		}

		// kalau ganti lecturer (bukan item pertama), sisip 1 baris kosong
		if index != 0 && (curLecturer != prevLecturer || curAcademicManager != prevAcademicManager) {
			if curAcademicManager != prevAcademicManager {
				f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), amTotalRealFee)
				f.SetCellStyle(sheetName, fmt.Sprintf("M%v", row), fmt.Sprintf("M%v", row), currencyBoldStyle)
				amTotalRealFee = 0
			}
			row++ // spare 1 row kosong
		}

		ProgramFeePerMeeting, _ := item.ProgramFeePerMeeting.Float64()
		FullFee, _ := item.FullFee.Float64()
		RealFee, _ := item.RealFee.Float64()

		f.SetCellValue(sheetName, fmt.Sprintf("R%v", row), item.RegistrationID)
		f.SetCellValue(sheetName, fmt.Sprintf("A%v", row), seq)
		if item.AcademicManagerName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), *item.AcademicManagerName)
		}
		if item.LecturerName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), *item.LecturerName)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), item.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%v", row), item.ProgramName)

		f.SetCellValue(sheetName, fmt.Sprintf("F%v", row), item.ProgramMeetings)
		f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), ProgramFeePerMeeting)
		f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), FullFee)
		f.SetCellFormula(sheetName, fmt.Sprintf("I%v", row), fmt.Sprintf("=F%v*G%v", row, row))
		if item.IsFullFee {
			f.SetCellValue(sheetName, fmt.Sprintf("J%v", row), "full")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("J%v", row), "tidak full")
		}
		if item.FL != nil {
			FL, _ := item.FL.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("K%v", row), FL)
		}
		if item.NL != nil {
			NL, _ := item.NL.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("L%v", row), NL)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), RealFee)
		if item.Notes != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("N%v", row), *item.Notes)
		}
		if item.MentorDetailFeeUsed != nil {
			MentorDetailFeeUsed, _ := item.MentorDetailFeeUsed.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("O%v", row), MentorDetailFeeUsed)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("P%v", row), item.AccquisitionRights)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%v", row), item.MarketerName)

		amTotalRealFee += RealFee

		// update prev lecturer & pindah ke baris berikutnya
		prevLecturer = curLecturer
		prevAcademicManager = curAcademicManager
		row++
		seq++
	}

	f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), amTotalRealFee)
	f.SetCellStyle(sheetName, fmt.Sprintf("M%v", row), fmt.Sprintf("M%v", row), currencyBoldStyle)

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-rekap-gaji-mentor-%v.xlsx", tmstmp)
	filepath := "./" + filename

	// save the file
	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename
	return resp, nil
}
