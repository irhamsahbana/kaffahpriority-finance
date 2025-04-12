package service

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

var _ ports.ReportService = &reportService{}

type reportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) *reportService {
	return &reportService{
		repo: repo,
	}
}

func (s *reportService) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	return s.repo.CreateTemplate(ctx, req)
}

func (s *reportService) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	return s.repo.UpdateTemplate(ctx, req)
}

func (s *reportService) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	return s.repo.GetTemplates(ctx, req)
}

func (s *reportService) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	return s.repo.GetTemplate(ctx, req)
}

func (s *reportService) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	return s.repo.CreateRegistrations(ctx, req)
}

func (s *reportService) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	return s.repo.CopyRegistrations(ctx, req)
}

func (s *reportService) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	return s.repo.UpdateRegistration(ctx, req)
}

func (s *reportService) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	return s.repo.GetRegistrations(ctx, req)
}

func (s *reportService) GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error) {
	return s.repo.GetRegistration(ctx, req)
}

func (s *reportService) GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error) {
	return s.repo.GetSummaries(ctx, req)
}

func (s *reportService) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	return s.repo.GetLecturerPrograms(ctx, req)
}

func (s *reportService) GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	return s.repo.GetRegistrationsPerLecturer(ctx, req)
}

func (s *reportService) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	return s.repo.GetLecturersWages(ctx, req)
}

func (s *reportService) GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error) {
	return s.repo.GetLecturersWagesAggregate(ctx, req)
}

func (s *reportService) UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error {
	return s.repo.UpdateLecturersWage(ctx, req)
}

func (s *reportService) DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error {
	return s.repo.DistributeHRFee(ctx, req)
}

func (s *reportService) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	return s.repo.UseHRfeeForLecturer(ctx, req)
}

func (s *reportService) GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error) {
	return s.repo.GetAcquisitionRightsAggregate(ctx, req)
}

func (s *reportService) GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error) {
	resp, err := s.repo.GetExportedRegistrations(ctx, req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	style, err := f.NewStyle(&excelize.Style{
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
			Color:   []string{"#FFFFFF"},
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

	numberFormat, err := f.NewStyle(&excelize.Style{
		NumFmt: 3,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("service::GetExportedRegistrations - error creating number format")
		return nil, err
	}

	f.NewSheet(sheetName)

	// column width
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 20)
	f.SetColWidth(sheetName, "G", "G", 20)
	f.SetColWidth(sheetName, "H", "H", 20)
	f.SetColWidth(sheetName, "I", "I", 20)
	f.SetColWidth(sheetName, "J", "J", 20)
	f.SetColWidth(sheetName, "K", "K", 20)
	f.SetColWidth(sheetName, "L", "L", 20)
	f.SetColWidth(sheetName, "M", "M", 20)
	f.SetColWidth(sheetName, "N", "N", 20)
	f.SetColWidth(sheetName, "O", "O", 20)

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

	f.SetCellStyle(sheetName, "A1", "O5", style)

	lastHariTanggal := ""
	lastRow := 5

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
		hariTanggal := paidAt.Format("Monday, 02/01/2006")
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
		if item.OverpaymentFee == nil {
			f.SetCellValue(sheetName, fmt.Sprintf("K%v", row), "")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("K%v", row), item.OverpaymentFee)
		}
		// CLOSINGAN
		// -- REWARD
		if item.ClosingFeeForReward == nil {
			f.SetCellValue(sheetName, fmt.Sprintf("L%v", row), "")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("L%v", row), item.ClosingFeeForReward)
		}
		// -- KEEP KANTOR
		if item.ClosingFeeForOffice == nil {
			f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), "")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("M%v", row), item.ClosingFeeForOffice)
		}
		// LABA
		f.SetCellValue(sheetName, fmt.Sprintf("N%v", row), item.Profit)

		// styling for number
		f.SetCellStyle(sheetName, fmt.Sprintf("G%v", row), fmt.Sprintf("N%v", row), numberFormat)
	}

	// PENYETORAN
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+1), "MENTOR")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+2), "KELEBIHAN")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+3), "ASET MARKETING")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+4), "HADIAH MARKETING")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+5), "REWARD")
	f.SetCellValue(sheetName, fmt.Sprintf("C%v", lastRow+6), "LABA")

	f.SetCellStyle(sheetName, fmt.Sprintf("C%v", lastRow+1), fmt.Sprintf("C%v", lastRow+6), penyetoranStyle)

	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+1), resp.Summary.TotalHrFee.String())
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+2), resp.Summary.TotalOverpaymentFee.String())
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+3), resp.Summary.TotalMarketerCommission.String())
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+4), resp.Summary.TotalMarketerGifts.String())
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+5), resp.Summary.TotalClosingFeeForReward.String())
	f.SetCellValue(sheetName, fmt.Sprintf("O%v", lastRow+6), resp.Summary.TotalProfit.String())

	f.SetCellStyle(sheetName, fmt.Sprintf("O%v", lastRow+1), fmt.Sprintf("O%v", lastRow+6), numberFormat)

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
	req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (
	*entity.GetExportedRegistrationsForCFO2MonthlyResp, error,
) {

	return nil, nil
}
