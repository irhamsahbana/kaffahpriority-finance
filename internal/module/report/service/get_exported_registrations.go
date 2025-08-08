package service

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func (s *reportService) GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error) {
	resp, err := s.repo.GetExportedRegistrations(ctx, req)
	if err != nil {
		return nil, err
	}

	var (
		fnName    = "service::GetExportedRegistrations"
		sheetName = "Sheet1"
	)

	f := excelize.NewFile()

	headerStyle, _ := newHeaderStyle(f, "#FFFF00", false)
	penyetoranStyle, _ := newPenyetoranStyle(f)
	numberFormatStyle, _ := newNumberFormatStyle(f)

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
			log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
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
			log.Error().Err(err).Any("req", req).Msgf("%s - error parsing date", fnName)
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
		log.Error().Err(err).Any("req", req).Msgf("%s - error saving file", fnName)
		return nil, err
	}

	resp.FilePath = filepath
	resp.FileName = filename
	return resp, nil
}

// newPenyetoranStyle membuat style teks tebal kiri untuk label PENYETORAN
// f: pointer ke excelize.File yang aktif
func newPenyetoranStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
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
}
