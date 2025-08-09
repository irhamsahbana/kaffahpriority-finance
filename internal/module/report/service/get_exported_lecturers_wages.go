package service

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func (s *reportService) GetExportedLecturersWages(ctx context.Context, req *entity.GetExportedLecturersWagesReq) (*entity.GetExportedLecturersWagesResp, error) {
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
	f.SetCellValue(sheetName, "N1", "Keterangan")
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
	for index, item := range resp.Items {
		no := index + 2

		ProgramFeePerMeeting, _ := item.ProgramFeePerMeeting.Float64()
		FullFee, _ := item.FullFee.Float64()
		InitialFee, _ := item.InitialFee.Float64()
		RealFee, _ := item.RealFee.Float64()

		f.SetCellValue(sheetName, fmt.Sprintf("R%v", no), item.RegistrationId)
		f.SetCellValue(sheetName, fmt.Sprintf("A%v", no), no)
		if item.AcademicManagerName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", no), *item.AcademicManagerName)
		}
		if item.LecturerName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%v", no), *item.LecturerName)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", no), item.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%v", no), item.ProgramName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%v", no), item.ProgramMeetings)
		f.SetCellValue(sheetName, fmt.Sprintf("G%v", no), ProgramFeePerMeeting)
		f.SetCellValue(sheetName, fmt.Sprintf("H%v", no), FullFee)
		f.SetCellValue(sheetName, fmt.Sprintf("I%v", no), InitialFee)
		if item.IsFullFee {
			f.SetCellValue(sheetName, fmt.Sprintf("J%v", no), "full")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("J%v", no), "tidak full")
		}
		if item.FL != nil {
			FL, _ := item.FL.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("K%v", no), FL)
		}
		if item.NL != nil {
			NL, _ := item.NL.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("L%v", no), NL)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("M%v", no), RealFee)
		if item.Notes != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("N%v", no), *item.Notes)
		}
		if item.MentorDetailFeeUsed != nil {
			MentorDetailFeeUsed, _ := item.MentorDetailFeeUsed.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("O%v", no), MentorDetailFeeUsed)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("P%v", no), item.AccquisitionRights)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%v", no), item.MarketerName)
	}

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
