package service

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func (s *reportService) GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error) {
	resp, err := s.repo.GetExportedRegistrations(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"},
			Pattern: 1,
		},
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating style")
		return nil, err
	}

	penyetoranStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating style")
		return nil, err
	}

	numberFormatStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 3,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating number format")
		return nil, err
	}

	f.NewSheet(sheetName)

	// column width
	f.SetColWidth(sheetName, "B", "O", 20)

	f.SetCellValue(sheetName, "A1", "JURNAL KEUANGAN CFO 1 KP")
	f.MergeCell(sheetName, "A1", "O2")

	f.SetCellValue(sheetName, "A3", "NO")
	f.MergeCell(sheetName, "A3", "A5")

	f.SetCellValue(sheetName, "B3", "HARI/TANGGAL")
	f.MergeCell(sheetName, "B3", "B5")

	f.SetCellValue(sheetName, "C3", "NIS")
	f.MergeCell(sheetName, "C3", "C5")

	f.SetCellValue(sheetName, "D3", "KETERANGAN")
	f.MergeCell(sheetName, "D3", "F4")
	f.SetCellValue(sheetName, "D5", "NAMA SANTRI")
	f.SetCellValue(sheetName, "E5", "MARKETER (MKP)")
	f.SetCellValue(sheetName, "F5", "PROGRAM")

	f.SetCellValue(sheetName, "G3", "DEBET (PEMASUKAN)")
	f.MergeCell(sheetName, "G3", "G5")

	f.SetCellValue(sheetName, "H3", "KREDIT (PENGELUARAN)")
	f.MergeCell(sheetName, "H3", "M3")
	f.SetCellValue(sheetName, "H4", "KOMISI/ASET MARKETING")
	f.MergeCell(sheetName, "H4", "H5")
	f.SetCellValue(sheetName, "I4", "HADIAH MARKETING")
	f.MergeCell(sheetName, "I4", "I5")
	f.SetCellValue(sheetName, "J4", "MENTOR + SDM + MS")
	f.MergeCell(sheetName, "J4", "J5")
	f.SetCellValue(sheetName, "K4", "KELEBIHAN")
	f.MergeCell(sheetName, "K4", "K5")
	f.SetCellValue(sheetName, "L4", "CLOSINGAN")
	f.MergeCell(sheetName, "L4", "M4")
	f.SetCellValue(sheetName, "L5", "REWARD")
	f.SetCellValue(sheetName, "M5", "KEEP KANTOR")

	f.SetCellValue(sheetName, "N3", "LABA")
	f.MergeCell(sheetName, "N3", "N5")

	f.SetCellValue(sheetName, "O3", "PENYETORAN")
	f.MergeCell(sheetName, "O3", "O5")

	f.SetCellStyle(sheetName, "A1", "O5", headerStyle)

	lastHariTanggal := ""
	lastRow := 5
	location, _ := time.LoadLocation(req.Timezone)

	if len(resp.Items) == 0 {
		tmstmp := time.Now().Unix()
		filename := fmt.Sprintf("laporan-keuangan-cfo-1-%v.xlsx", tmstmp)
		filepath := "./" + filename

		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+1), "MENTOR")
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+2), "KELEBIHAN")
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+3), "ASET MARKETING")
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+4), "HADIAH MARKETING")
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+5), "REWARD")
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+6), "LABA")

		// save the file
		if err := f.SaveAs(filepath); err != nil {
			log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error saving file")
			return nil, err
		}

		resp.FilePath = filepath
		resp.FileName = filename
		return resp, nil
	}

	for i, item := range resp.Items {
		row := i + 6 // start from row 6
		lastRow = row

		//hariTanggal format: Rabu, 04/12/2024
		// item.PaidAt is still in string that came from sql without modification so we need convert to time.Time
		paidAt, err := time.Parse("2006-01-02T15:04:05Z", item.PaidAt)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error parsing date")
			return nil, err
		}

		// hari tanggal menggunakan bahasa indonesia
		paidAt = paidAt.In(location)
		hariTanggal := hariTanggalString(paidAt)

		if lastHariTanggal != "" && lastHariTanggal == hariTanggal {
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), "")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), hariTanggal)
			lastHariTanggal = hariTanggal
		}

		// NIS
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), item.StudentIdentifier)
		// NAMA SANTRI
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), item.StudentName)
		// MARKETER (MKP)
		f.SetCellValue(sheetName, fmt.Sprintf("E%v", row), item.MarketerName)
		// PROGRAM
		f.SetCellValue(sheetName, fmt.Sprintf("F%v", row), item.ProgramName)
		// DEBET (PEMASUKAN)
		f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), item.MonthlyFee)
		// KREDIT (PENGELUARAN)
		// -- KOMISI/ASET MARKETING
		f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), item.MarketerCommissionFee)
		// -- HADIAH MARKETING
		f.SetCellValue(sheetName, fmt.Sprintf("I%v", row), item.MarketerGiftsFee)
		// -- MENTOR + SDM + MS
		f.SetCellValue(sheetName, fmt.Sprintf("J%v", row), item.HRFee)
		// -- KELEBIHAN
		if item.OverpaymentFee != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("K%v", row), *item.OverpaymentFee)
		}
		// CLOSINGAN
		// -- REWARD
		if item.ClosingFeeForReward != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("L%v", row), *item.ClosingFeeForReward)
		}
		// -- KEEP KANTOR
		if item.ClosingFeeForOffice != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), *item.ClosingFeeForOffice)
		}
		// LABA
		f.SetCellValue(sheetName, fmt.Sprintf("N%v", row), item.Profit)

		// styling for number
		f.SetCellStyle(sheetName, fmt.Sprintf("G%v", row), fmt.Sprintf("N%v", row), numberFormatStyle)
	}

	// PENYETORAN
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+1), "MENTOR")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+2), "KELEBIHAN")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+3), "ASET MARKETING")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+4), "HADIAH MARKETING")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+5), "REWARD")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+6), "LABA")

	f.SetCellStyle(sheetName, fmt.Sprintf("C%v", lastRow+1), fmt.Sprintf("C%v", lastRow+6), penyetoranStyle)

	totalHRFeeFloat, _ := resp.Summary.TotalHrFee.Float64()
	totalOverpaymentFeeFloat, _ := resp.Summary.TotalOverpaymentFee.Float64()
	totalMarketerCommissionFloat, _ := resp.Summary.TotalMarketerCommission.Float64()
	totalMarketerGiftsFloat, _ := resp.Summary.TotalMarketerGifts.Float64()
	totalClosingFeeForRewardFloat, _ := resp.Summary.TotalClosingFeeForReward.Float64()
	totalProfitFloat, _ := resp.Summary.TotalProfit.Float64()

	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+1), totalHRFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+2), totalOverpaymentFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+3), totalMarketerCommissionFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+4), totalMarketerGiftsFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+5), totalClosingFeeForRewardFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+6), totalProfitFloat)

	f.SetCellStyle(sheetName, fmt.Sprintf("O%v", lastRow+1), fmt.Sprintf("O%v", lastRow+6), numberFormatStyle)

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-cfo-1-%v.xlsx", tmstmp)
	filepath := "./" + filename

	// save the file
	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error saving file")
		log.Debug().Any("lastRow", lastRow).Msg("service::GetExportedRegistrations - last row")
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename
	return resp, nil
}

func (s *reportService) GetExportedRegistrationsForCFO2Monthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	resp, err := s.repo.GetExportedRegistrationsForCFO2Monthly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	HeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"},
			Pattern: 1,
		},
		NumFmt: 3,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating style")
		return nil, err
	}

	numberFormatStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 3,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating number format")
		return nil, err
	}

	unusedStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating unused style")
		return nil, err
	}

	timeStart, _ := time.Parse("2006-01-02", req.PaidAtFrom)
	timeEnd, _ := time.Parse("2006-01-02", req.PaidAtTo)
	location, _ := time.LoadLocation(req.Timezone)

	timeStartLocal := timeStart.In(location)
	timeEndLocal := timeEnd.In(location)
	// format DD/MM/YYYY
	timeStartLocalFormatted := timeStartLocal.Format("02/01/2006")
	timeEndLocalFormatted := timeEndLocal.Format("02/01/2006")

	f.SetCellValue(sheetName, "A1", "LAPORAN KEUANGAN AKADEMIK "+timeStartLocalFormatted+" - "+timeEndLocalFormatted)
	f.MergeCell(sheetName, "A1", "I2")

	f.SetCellValue(sheetName, "A3", "TANGGAL")
	f.MergeCell(sheetName, "A3", "A4")
	f.SetCellValue(sheetName, "B3", "S")
	f.MergeCell(sheetName, "B3", "B4")
	f.SetCellValue(sheetName, "C3", "NAMA")
	f.MergeCell(sheetName, "C3", "C4")
	f.SetCellValue(sheetName, "D3", "PROGRAM")
	f.MergeCell(sheetName, "D3", "D4")
	f.SetCellValue(sheetName, "E3", "RINCIAN")
	f.MergeCell(sheetName, "E3", "G3")
	f.SetCellValue(sheetName, "E4", "MENTOR")
	f.SetCellValue(sheetName, "F4", "SDM")
	f.SetCellValue(sheetName, "G4", "TOTAL")
	f.SetCellValue(sheetName, "H3", "DEBET (PEMASUKAN)")
	f.MergeCell(sheetName, "H3", "H4")
	f.SetCellValue(sheetName, "I3", "KREDIT (PENGELUARAN)")
	f.MergeCell(sheetName, "I3", "I4")

	f.SetCellStyle(sheetName, "A1", "I4", HeaderStyle)
	f.SetColWidth(sheetName, "A", "J", 20)

	lastRow := 4
	lastHariTanggal := ""
	location, _ = time.LoadLocation(req.Timezone)
	totalStudentParticipant := 0
	totalMentorFee := decimal.NewFromFloat(0)
	totalHRFee := decimal.NewFromFloat(0)
	totalIncome := decimal.NewFromFloat(0)
	totalOutcome := decimal.NewFromFloat(0)

	// if there is no data, we need to create a file with the header only
	if len(resp.Items) == 0 {
		tmstmp := time.Now().Unix()
		filename := fmt.Sprintf("laporan-keuangan-cfo-2-bulanan-%v.xlsx", tmstmp)
		filepath := "./" + filename

		if err := f.SaveAs(filepath); err != nil {
			log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error saving file")
			return nil, err
		}

		resp.FilePath = filepath
		resp.FileName = filename
		return resp, nil
	}

	for i, item := range resp.Items {
		row := i + 5 // start from row 5
		lastRow = row

		//hariTanggal format: Rabu, 04/12/2024
		// item.PaidAt is still in string that came from sql without modification so we need convert to time.Time
		paidAt, err := time.Parse("2006-01-02T15:04:05Z", item.PaidAt)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error parsing date")
			return nil, err
		}

		// hari tanggal menggunakan bahasa indonesia
		paidAt = paidAt.In(location)
		hariTanggal := hariTanggalString(paidAt)

		if lastHariTanggal != "" && lastHariTanggal == hariTanggal {
			f.SetCellValue(sheetName, fmt.Sprintf("A%v", row), "")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("A%v", row), hariTanggal)
			lastHariTanggal = hariTanggal
		}

		studentParticipant := 1
		totalStudentParticipant++
		for range item.Students {
			studentParticipant++
			totalStudentParticipant++
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), studentParticipant)

		// NAMA
		f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), item.StudentName)
		// PROGRAM
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), item.ProgramName)
		// RINCIAN
		// -- MENTOR
		if item.HRFeeForMentor != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%v", row), *item.HRFeeForMentor)
		}
		// -- SDM
		if item.HRFeeForHR != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("F%v", row), *item.HRFeeForHR)
		}
		// -- TOTAL
		if item.HRFeeForHR != nil && item.HRFeeForMentor != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), *item.HRFeeForHR+*item.HRFeeForMentor)
			f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), *item.HRFeeForHR+*item.HRFeeForMentor) // DEBET (PEMASUKAN)
			totalMentorFee = totalMentorFee.Add(decimal.NewFromFloat(*item.HRFeeForMentor))
			totalHRFee = totalHRFee.Add(decimal.NewFromFloat(*item.HRFeeForHR))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.HRFeeForHR + *item.HRFeeForMentor))
		} else if item.HRFeeForHR != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), *item.HRFeeForHR)
			f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), *item.HRFeeForHR) // DEBET (PEMASUKAN)
			totalHRFee = totalHRFee.Add(decimal.NewFromFloat(*item.HRFeeForHR))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.HRFeeForHR))
		} else if item.HRFeeForMentor != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), *item.HRFeeForMentor)
			f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), *item.HRFeeForMentor) // DEBET (PEMASUKAN)
			totalMentorFee = totalMentorFee.Add(decimal.NewFromFloat(*item.HRFeeForMentor))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.HRFeeForMentor))
		}

		if item.IsUnused {
			f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("I%v", row), unusedStyle)
		}
	}

	f.SetCellStyle(sheetName, "E5", fmt.Sprintf("H%v", lastRow+1), numberFormatStyle)

	totalMentorFeeFloat, _ := totalMentorFee.Float64()
	totalHRFeeFloat, _ := totalHRFee.Float64()
	totalIncomeFloat, _ := totalIncome.Float64()
	totalOutcomeFloat, _ := totalOutcome.Float64()
	totalRemainingFloat, _ := totalIncome.Sub(totalOutcome).Float64()

	f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow+3), resp.TotalITP)
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+3), "ITP")

	f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow+2), totalStudentParticipant)
	f.SetCellValue(sheetName, fmt.Sprintf("E%v", lastRow+2), totalMentorFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("F%v", lastRow+2), totalHRFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("H%v", lastRow+2), totalIncomeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("I%v", lastRow+2), totalOutcomeFloat)

	f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow+1), "TOTAL SALDO")
	f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow+2), totalRemainingFloat)

	f.SetCellStyle(sheetName, fmt.Sprintf("A%v", lastRow+2), fmt.Sprintf("J%v", lastRow+2), HeaderStyle)

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-cfo-2-bulanan-%v.xlsx", tmstmp)
	filepath := "./" + filename

	// save the file
	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error saving file")
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename

	return resp, nil
}

func (s *reportService) GetExportedRegistrationsForCFO2Yearly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2YearlyReq) (*entity.GetExportedRegistrationsForCFO2YearlyResp, error) {
	resp, err := s.repo.GetExportedRegistrationsForCFO2Yearly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	borderStyle := []excelize.Border{
		{
			Type:  "left",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "top",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "bottom",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "right",
			Color: "#000000",
			Style: 1,
		},
	}

	HeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: borderStyle,
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"},
			Pattern: 1,
		},
		NumFmt: 3,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrationsForCFO2Yearly - error creating style")
		return nil, err
	}

	bodyStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 12,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: borderStyle,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrationsForCFO2Yearly - error creating body style")
		return nil, err
	}

	// Headering START
	f.SetCellValue(sheetName, "A1", "LAPORAN KEUANGAN AKADEMIK "+req.PaidAtYear)
	f.MergeCell(sheetName, "A1", "R1")

	f.SetCellValue(sheetName, "A2", "NO")
	f.SetCellValue(sheetName, "B2", "MENTOR")
	f.SetCellValue(sheetName, "C2", "NO")
	f.SetCellValue(sheetName, "D2", "NAMA SANTRI")
	f.SetCellValue(sheetName, "E2", "PROGRAM")
	f.SetCellValue(sheetName, "F2", "MANAGER AKADEMIK")

	// looping through months in the year, with example: JAN, FEB, MAR, APR, MAY, JUN, JUL, AUG, SEP, OCT, NOV, DEC
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for i, month := range months {
		f.SetCellValue(sheetName, fmt.Sprintf("%s2", string(rune('G'+i))), month)
	}

	f.SetCellStyle(sheetName, "A1", "R2", HeaderStyle)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 20)
	// Headering END

	// section for data
	lastRow := 2

	for _, item := range resp.Items {
		lastRow++
		f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow), item.LecturerName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", lastRow), item.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%v", lastRow), item.ProgramName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%v", lastRow), item.AcademicManagerName)

		for _, month := range item.Months {
			monthNumber := month.PaidAtMonth

			if month.HRFeeForMentor != nil {
				f.SetCellValue(sheetName, fmt.Sprintf("%s%v", string(rune('G'+monthNumber-1)), lastRow), *month.HRFeeForMentor)
			}

			if month.Notes != nil {
				f.SetCellValue(sheetName, fmt.Sprintf("%s%v", string(rune('G'+monthNumber-1)), lastRow), *month.Notes)
			}
		}
	}

	// set the style for the body
	for i := 3; i <= lastRow; i++ {
		f.SetCellStyle(sheetName, fmt.Sprintf("A%v", i), fmt.Sprintf("R%v", i), bodyStyle)
	}

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-cfo-2-tahunan-%v.xlsx", tmstmp)
	filepath := "./" + filename

	// save the file
	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error saving file")
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename

	return resp, nil

}

func (s *reportService) GetExportedRegistrationsForWageRecapMonthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error) {
	resp, err := s.repo.GetExportedRegistrationsForWageRecapMonthly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	borderStyle := []excelize.Border{
		{
			Type:  "left",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "top",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "bottom",
			Color: "#000000",
			Style: 1,
		},
		{
			Type:  "right",
			Color: "#000000",
			Style: 1,
		},
	}

	HeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: borderStyle,
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

	// Headering START
	f.SetCellValue(sheetName, "A1", "NO")
	f.SetCellValue(sheetName, "B1", "NAMA")
	f.SetCellValue(sheetName, "C1", "NO")
	f.SetCellValue(sheetName, "D1", "NAMA")
	f.SetCellValue(sheetName, "E1", "PROGRAM")
	f.SetCellValue(sheetName, "F1", "JUMLAH")
	f.SetCellValue(sheetName, "G1", "HITUNGAN")
	f.SetCellValue(sheetName, "H1", "UJROH FULL")
	f.SetCellValue(sheetName, "I1", "UJROH AWAL")
	f.SetCellValue(sheetName, "J1", "FL")
	f.SetCellValue(sheetName, "K1", "NL")
	f.SetCellValue(sheetName, "L1", "UJROH REAL")
	f.SetCellValue(sheetName, "M1", "KETERANGAN")
	f.SetCellValue(sheetName, "N1", "KEEP GAJI")
	f.SetCellValue(sheetName, "O1", "ANGKA")
	f.SetCellValue(sheetName, "P1", "PENASEHAT AKADEMIK")

	f.SetCellStyle(sheetName, "A1", "Q1", HeaderStyle)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "G", "I", 20)
	f.SetColWidth(sheetName, "L", "L", 20)
	f.SetColWidth(sheetName, "N", "O", 20)
	f.SetColWidth(sheetName, "P", "P", 20)
	// Headering END

	lastRow := 1

	for _, academicManager := range resp.Items {
		lastRow++
		f.SetCellValue(sheetName, fmt.Sprintf("A%v", lastRow), academicManager.Name)
		f.MergeCell(sheetName, fmt.Sprintf("A%v", lastRow), fmt.Sprintf("P%v", lastRow+1))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%v", lastRow), fmt.Sprintf("P%v", lastRow+1), academicManagerNameStyle)
		lastRow++

		for lecturerIndex, lecturer := range academicManager.Items {
			lastRow++
			f.SetCellValue(sheetName, fmt.Sprintf("A%v", lastRow), lecturerIndex+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow), lecturer.Name)
			for templateIndex, template := range lecturer.Items {
				f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow), templateIndex+1)
				f.SetCellValue(sheetName, fmt.Sprintf("D%v", lastRow), template.StudentName)
				f.SetCellValue(sheetName, fmt.Sprintf("E%v", lastRow), template.ProgramName)
				f.SetCellValue(sheetName, fmt.Sprintf("P%v", lastRow), template.MarketerName)
				if template.Data != nil {
					data := template.Data
					f.SetCellValue(sheetName, fmt.Sprintf("F%v", lastRow), data.ProgramMeetings)
					f.SetCellValue(sheetName, fmt.Sprintf("G%v", lastRow), data.ProgramFeePerMeeting)
					f.SetCellValue(sheetName, fmt.Sprintf("H%v", lastRow), data.FullFee)
					if data.InitialFee != nil {
						f.SetCellValue(sheetName, fmt.Sprintf("I%v", lastRow), *data.InitialFee)
					}
					if data.FL != nil {
						f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow), *data.FL)
					}
					if data.NL != nil {
						f.SetCellValue(sheetName, fmt.Sprintf("K%v", lastRow), *data.NL)
					}
					f.SetCellValue(sheetName, fmt.Sprintf("L%v", lastRow), data.RealFee)
					f.SetCellValue(sheetName, fmt.Sprintf("N%v", lastRow), data.MentorDetailFeeUsed)
					f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow), data.AcquisitionRights)
				}

				if templateIndex+1 != len(lecturer.Items) { // not the last item in the lecturer.Items
					lastRow++
				}
			}
		}
	}

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-rekap-gaji-bulanan-%s-%v.xlsx", req.Month, tmstmp)
	filepath := "./" + filename

	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrationsForX - error saving file")
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename

	return resp, nil
}

func hariTanggalString(htd time.Time) string {
	hariTanggal := htd.Format("Monday, 02/01/2006")
	if strings.Contains(hariTanggal, "Monday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Monday", "Senin")
	} else if strings.Contains(hariTanggal, "Tuesday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Tuesday", "Selasa")
	} else if strings.Contains(hariTanggal, "Wednesday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Wednesday", "Rabu")
	} else if strings.Contains(hariTanggal, "Thursday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Thursday", "Kamis")
	} else if strings.Contains(hariTanggal, "Friday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Friday", "Jumat")
	} else if strings.Contains(hariTanggal, "Saturday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Saturday", "Sabtu")
	} else if strings.Contains(hariTanggal, "Sunday") {
		hariTanggal = strings.ReplaceAll(hariTanggal, "Sunday", "Minggu")
	}

	return hariTanggal
}
