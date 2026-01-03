package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func (s *reportService) GetExportedRegistrationsForCFO2Monthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetExportedRegistrationsForCFO2Monthly")
	defer span.End()

	resp, err := s.repo.GetExportedRegistrationsForCFO2Monthly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()

	var (
		fnName    = "service::GetExportedRegistrationsForCFO2Monthly"
		sheetName = "Sheet1"
	)

	// Header kuning dengan NumFmt
	HeaderStyle, _ := newHeaderStyle(f, "#FFFF00", true)

	// Style angka NumFmt:3
	numberFormatStyle, _ := newNumberFormatStyle(f)

	// Style penandaan unused
	unusedStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
	})

	timeStart, _ := time.Parse("2006-01-02", req.PaidAtFrom)
	timeEnd, _ := time.Parse("2006-01-02", req.PaidAtTo)
	location, _ := time.LoadLocation(req.Timezone)

	timeStartLocal := timeStart.In(location)
	timeEndLocal := timeEnd.In(location)
	// format DD/MM/YYYY
	timeStartLocalFormatted := timeStartLocal.Format("02/01/2006")
	timeEndLocalFormatted := timeEndLocal.Format("02/01/2006")

	f.SetCellValue(sheetName, "A1", "LAPORAN KEUANGAN AKADEMIK "+timeStartLocalFormatted+" - "+timeEndLocalFormatted)
	f.MergeCell(sheetName, "A1", "J2")

	f.SetCellValue(sheetName, "A3", "TANGGAL")
	f.MergeCell(sheetName, "A3", "A4")
	f.SetCellValue(sheetName, "B3", "HAK")
	f.MergeCell(sheetName, "B3", "B4")
	f.SetCellValue(sheetName, "C3", "NAMA")
	f.MergeCell(sheetName, "C3", "C4")
	f.SetCellValue(sheetName, "D3", "PROGRAM")
	f.MergeCell(sheetName, "D3", "D4")
	f.SetCellValue(sheetName, "E3", "RINCIAN")
	f.MergeCell(sheetName, "E3", "G3")
	f.SetCellValue(sheetName, "E4", "MENTOR")
	f.SetCellValue(sheetName, "F4", "SDM")
	f.SetCellValue(sheetName, "G4", "KELEBIHAN")

	f.SetCellValue(sheetName, "H4", "TOTAL")
	f.SetCellValue(sheetName, "I3", "DEBET (PEMASUKAN)")
	f.MergeCell(sheetName, "I3", "I4")
	f.SetCellValue(sheetName, "J3", "KREDIT (PENGELUARAN)")
	f.MergeCell(sheetName, "J3", "J4")

	f.SetCellStyle(sheetName, "A1", "J4", HeaderStyle)
	f.SetColWidth(sheetName, "A", "J", 20)

	lastRow := 4
	lastHariTanggal := ""
	location, _ = time.LoadLocation(req.Timezone)
	// totalStudentParticipant := 0
	totalAcquisitionRights := 0
	totalMentorFee := decimal.NewFromFloat(0)
	totalHRFee := decimal.NewFromFloat(0)
	totalOverpaymentFee := decimal.NewFromFloat(0)
	totalIncome := decimal.NewFromFloat(0)
	totalOutcome := decimal.NewFromFloat(0)

	// if there is no data, we need to create a file with the header only
	if len(resp.Items) == 0 {
		tmstmp := time.Now().Unix()
		filename := fmt.Sprintf("laporan-keuangan-cfo-2-bulanan-%v.xlsx", tmstmp)
		filepath := "./" + filename

		if err := f.SaveAs(filepath); err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
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
			log.Error().Err(err).Any("req", req).Msgf("%s - error parsing date", fnName)
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

		// studentParticipant := 1
		// totalStudentParticipant++
		// for range item.Students {
		// 	studentParticipant++
		// 	totalStudentParticipant++
		// }
		// f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), studentParticipant)
		totalAcquisitionRights += item.AcquisitionRights
		f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), item.AcquisitionRights)

		// NAMA
		// allocatedAt is pointer to string, so we need to check if it is nil then only get year and month, not the day
		var allocatedAt string
		if item.AllocatedAt != nil {
			allocatedAtTime, err := time.Parse("2006-01-02T15:04:05Z", *item.AllocatedAt)
			if err != nil {
				log.Error().Err(err).Any("req", req).Msgf("%s - error parsing allocated at date", fnName)
				return nil, err
			}
			allocatedAtTime = allocatedAtTime.In(location)
			// format "nama_bulan_indonesia" + tahun

			allocatedAt = allocatedAtTime.Format("January 2006")

			if strings.Contains(allocatedAt, "January") {
				allocatedAt = strings.Replace(allocatedAt, "January", "Januari", 1)
			} else if strings.Contains(allocatedAt, "February") {
				allocatedAt = strings.Replace(allocatedAt, "February", "Februari", 1)
			} else if strings.Contains(allocatedAt, "March") {
				allocatedAt = strings.Replace(allocatedAt, "March", "Maret", 1)
			} else if strings.Contains(allocatedAt, "April") {
				allocatedAt = strings.Replace(allocatedAt, "April", "April", 1)
			} else if strings.Contains(allocatedAt, "May") {
				allocatedAt = strings.Replace(allocatedAt, "May", "Mei", 1)
			} else if strings.Contains(allocatedAt, "June") {
				allocatedAt = strings.Replace(allocatedAt, "June", "Juni", 1)
			} else if strings.Contains(allocatedAt, "July") {
				allocatedAt = strings.Replace(allocatedAt, "July", "Juli", 1)
			} else if strings.Contains(allocatedAt, "August") {
				allocatedAt = strings.Replace(allocatedAt, "August", "Agustus", 1)
			} else if strings.Contains(allocatedAt, "September") {
				allocatedAt = strings.Replace(allocatedAt, "September", "September", 1)
			} else if strings.Contains(allocatedAt, "October") {
				allocatedAt = strings.Replace(allocatedAt, "October", "Oktober", 1)
			} else if strings.Contains(allocatedAt, "November") {
				allocatedAt = strings.Replace(allocatedAt, "November", "November", 1)
			} else if strings.Contains(allocatedAt, "December") {
				allocatedAt = strings.Replace(allocatedAt, "December", "Desember", 1)
			}

		} else {
			allocatedAt = "N/A"
		}

		f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), fmt.Sprintf("%s (%s)", item.StudentName, allocatedAt))
		// PROGRAM
		f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), item.ProgramName)
		// RINCIAN
		var totalForRow *decimal.Decimal
		// -- MENTOR
		if item.HRFeeForMentor != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%v", row), *item.HRFeeForMentor)
			totalMentorFee = totalMentorFee.Add(decimal.NewFromFloat(*item.HRFeeForMentor))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.HRFeeForMentor))

			t := decimal.NewFromFloat(0)
			totalForRow = &t
			*totalForRow = totalForRow.Add(decimal.NewFromFloat(*item.HRFeeForMentor))
		}
		// -- SDM
		if item.HRFeeForHR != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("F%v", row), *item.HRFeeForHR)
			totalHRFee = totalHRFee.Add(decimal.NewFromFloat(*item.HRFeeForHR))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.HRFeeForHR))

			if totalForRow == nil {
				t := decimal.NewFromFloat(0)
				totalForRow = &t
			}
			*totalForRow = totalForRow.Add(decimal.NewFromFloat(*item.HRFeeForHR))
		}
		// -- KELEBIHAN
		if item.OverpaymentFee != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%v", row), *item.OverpaymentFee)
			totalOverpaymentFee = totalOverpaymentFee.Add(decimal.NewFromFloat(*item.OverpaymentFee))
			totalIncome = totalIncome.Add(decimal.NewFromFloat(*item.OverpaymentFee))

			if totalForRow == nil {
				t := decimal.NewFromFloat(0)
				totalForRow = &t
			}
			*totalForRow = totalForRow.Add(decimal.NewFromFloat(*item.OverpaymentFee))
		}

		if totalForRow != nil {
			totalForRowFloat, _ := totalForRow.Float64()
			f.SetCellValue(sheetName, fmt.Sprintf("H%v", row), totalForRowFloat)
			f.SetCellValue(sheetName, fmt.Sprintf("I%v", row), totalForRowFloat) // DEBET (PEMASUKAN)
		}

		if item.IsUnused {
			f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("I%v", row), unusedStyle)
		}
	}

	f.SetCellStyle(sheetName, "E5", fmt.Sprintf("I%v", lastRow+1), numberFormatStyle)

	totalMentorFeeFloat, _ := totalMentorFee.Float64()
	totalHRFeeFloat, _ := totalHRFee.Float64()
	totalOverpaymentFeeFloat, _ := totalOverpaymentFee.Float64()
	totalIncomeFloat, _ := totalIncome.Float64()
	totalOutcomeFloat, _ := totalOutcome.Float64()
	totalRemainingFloat, _ := totalIncome.Sub(totalOutcome).Float64()

	f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow+3), resp.TotalITP)
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+3), "ITP")

	f.SetCellValue(sheetName, fmt.Sprintf("B%v", lastRow+2), totalAcquisitionRights)
	f.SetCellValue(sheetName, fmt.Sprintf("E%v", lastRow+2), totalMentorFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("F%v", lastRow+2), totalHRFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("G%v", lastRow+2), totalOverpaymentFeeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("I%v", lastRow+2), totalIncomeFloat)
	f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow+2), totalOutcomeFloat)

	f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow+1), "TOTAL SALDO")
	f.SetCellValue(sheetName, fmt.Sprintf("J%v", lastRow+2), totalRemainingFloat)

	f.SetCellStyle(sheetName, fmt.Sprintf("A%v", lastRow+2), fmt.Sprintf("J%v", lastRow+2), HeaderStyle)

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-cfo-2-bulanan-%v.xlsx", tmstmp)
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

func (s *reportService) GetExportedRegistrationsForCFO2Yearly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2YearlyReq) (*entity.GetExportedRegistrationsForCFO2YearlyResp, error) {
	resp, err := s.repo.GetExportedRegistrationsForCFO2Yearly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()

	var (
		fnName    = "service::GetExportedRegistrationsForCFO2Yearly"
		sheetName = "Sheet1"
	)

	// Header kuning dengan NumFmt
	HeaderStyle, _ := newHeaderStyle(f, "#FFFF00", true)

	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 12,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: defaultBorderStyle(),
	})

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
		log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename

	return resp, nil

}

func (s *reportService) GetExportedRegistrationsForWageRecapMonthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetExportedRegistrationsForWageRecapMonthly")
	defer span.End()

	resp, err := s.repo.GetExportedRegistrationsForWageRecapMonthly(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()

	var (
		fnName    = "service::GetExportedRegistrationsForWageRecapMonthly"
		sheetName = "Sheet1"
	)

	// Header biru muda dengan NumFmt
	HeaderStyle, _ := newHeaderStyle(f, "#3BFFF5", true)

	// Header biru muda dengan NumFmt rata kanan
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

	f.SetCellStyle(sheetName, "A1", "P1", HeaderStyle)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "G", "I", 20)
	f.SetColWidth(sheetName, "L", "L", 20)
	f.SetColWidth(sheetName, "M", "M", 20)
	f.SetColWidth(sheetName, "N", "O", 20)
	f.SetColWidth(sheetName, "P", "P", 20)
	// Headering END

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze: true,
		XSplit: 5,
		YSplit: 1,
	})

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

			var totalRealFee float64

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
					totalRealFee += data.RealFee
					if data.Notes != nil {
						f.SetCellValue(sheetName, fmt.Sprintf("M%v", lastRow), *data.Notes)
					}
					if data.MentorDetailFeeUsed != nil {
						f.SetCellValue(sheetName, fmt.Sprintf("N%v", lastRow), *data.MentorDetailFeeUsed)
					}
					f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow), data.AcquisitionRights)
				}

				if templateIndex+1 != len(lecturer.Items) { // not the last item in the lecturer.Items
					lastRow++
				}

				if templateIndex == len(lecturer.Items)-1 { // last item in the lecturer.Items
					lastRow++
					f.SetCellValue(sheetName, fmt.Sprintf("K%v", lastRow), "TOTAL")
					f.SetCellValue(sheetName, fmt.Sprintf("L%v", lastRow), totalRealFee)
					f.SetCellStyle(sheetName, fmt.Sprintf("K%v", lastRow), fmt.Sprintf("K%v", lastRow), HeaderStyle)
					f.SetCellStyle(sheetName, fmt.Sprintf("L%v", lastRow), fmt.Sprintf("L%v", lastRow), HeaderStyleRight)
					lastRow++ // add an extra row for the next lecturer
				}
			}
		}
	}

	// timestampe in unix
	tmstmp := time.Now().Unix()

	filename := fmt.Sprintf("laporan-keuangan-rekap-gaji-bulanan-%s-%v.xlsx", req.Month, tmstmp)
	filepath := "./" + filename

	if err := f.SaveAs(filepath); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
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

// =====================
// Helpers untuk styling
// =====================

// defaultBorderStyle mengembalikan border hitam tipis di keempat sisi
func defaultBorderStyle() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "#000000", Style: 1},
		{Type: "top", Color: "#000000", Style: 1},
		{Type: "bottom", Color: "#000000", Style: 1},
		{Type: "right", Color: "#000000", Style: 1},
	}
}

// newHeaderStyle membuat style header Excel dengan parameter warna fill dan opsi NumFmt
// f: pointer ke excelize.File yang aktif
// fillColor: warna background dalam format hex, mis. "#FFFF00"
// withNumFmt: jika true maka menambahkan NumFmt 3 ke style
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

// newNumberFormatStyle membuat style angka dengan NumFmt 3
// f: pointer ke excelize.File yang aktif
func newNumberFormatStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		NumFmt: 3,
	})
}
