package payroll

import (
	"codebase-app/internal/entity"
	"context"
)

type PayrollRepository interface {
	CreatePayrollRun(ctx context.Context, run *entity.PayrollRun) error
	CreatePayrollItem(ctx context.Context, req *entity.CreatePayrollItemReq) (*entity.CreatePayrollItemResp, error)
	CreatePayrollItems(ctx context.Context, items []entity.PayrollItem) error
	GetPayrollRuns(ctx context.Context, req *entity.GetPayrollRunsReq) ([]entity.PayrollRun, int, error)
	GetPayrollRun(ctx context.Context, id string) (*entity.PayrollRun, error)
	GetPayrollItems(ctx context.Context, payrollRunID string) ([]entity.PayrollItem, error)
	GetPayrollItem(ctx context.Context, id string) (*entity.PayrollItem, error)
	GetPayrollRunByPeriod(ctx context.Context, period string) (*entity.PayrollRun, error)
	GetLatestPayrollRun(ctx context.Context) (*entity.PayrollRun, error)
	GetPayrollItemsWithPagination(ctx context.Context, payrollRunID string, req *entity.GetPayrollItemsReq) ([]entity.PayrollItem, int, error)
	UpdatePayrollItem(ctx context.Context, req *entity.UpdatePayrollItemReq) error
	DeletePayrollItem(ctx context.Context, id string) error
	GetSourceDataForPayroll(ctx context.Context, period string, timezone string) ([]entity.PayrollItem, error)
	GeneratePayrollRun(ctx context.Context, period string, timezone string) (*entity.PayrollRun, error)
	GetPayrollRunDetail(ctx context.Context, id string) (*entity.GetPayrollRunDetailResp, error)
	GetAdditionalStudents(ctx context.Context, payrollItemIDs []string) (map[string][]entity.PayrollItemAdditionalStudent, error)
	GetPayrollItemsForExport(ctx context.Context, payrollRunID string) ([]entity.PayrollItem, error)
	BulkUpdatePayrollItems(ctx context.Context, req *entity.BulkUpdatePayrollItemReq) error
	ImportPayrollItems(ctx context.Context, req *entity.ImportPayrollItemsReq) (int64, error)
	GetPayrollItemsPeriodically(ctx context.Context, req *entity.GetPayrollItemsPeriodicallyReq) (*entity.GetPayrollItemsPeriodicallyResp, error)
	GetPayrollReportsYearly(ctx context.Context, req *entity.GetPayrollReportsYearlyReq) (*entity.PayrollReportsYearlyResp, error)
	ResetPayrollRunItems(ctx context.Context, runID string) (int64, error)
}

type PayrollService interface {
	CreatePayrollRun(ctx context.Context, req *entity.CreatePayrollRunReq) (*entity.CreatePayrollRunResp, error)
	CreatePayrollItem(ctx context.Context, req *entity.CreatePayrollItemReq) (*entity.CreatePayrollItemResp, error)
	GetPayrollRuns(ctx context.Context, req *entity.GetPayrollRunsReq) (*entity.GetPayrollRunsResp, error)
	GetPayrollRunDetail(ctx context.Context, req *entity.GetPayrollRunDetailReq) (*entity.GetPayrollRunDetailResp, error)
	GetPayrollItems(ctx context.Context, req *entity.GetPayrollItemsReq) (*entity.GetPayrollItemsResp, error)
	GetPayrollItemDetail(ctx context.Context, req *entity.GetPayrollItemDetailReq) (*entity.GetPayrollItemDetailResp, error)
	GetPayrollItemsPeriodically(ctx context.Context, req *entity.GetPayrollItemsPeriodicallyReq) (*entity.GetPayrollItemsPeriodicallyResp, error)
	UpdatePayrollItem(ctx context.Context, req *entity.UpdatePayrollItemReq) error
	BulkUpdatePayrollItems(ctx context.Context, req *entity.BulkUpdatePayrollItemReq) error
	ImportPayrollItems(ctx context.Context, req *entity.ImportPayrollItemsReq) (*entity.ImportPayrollItemsResp, error)
	DeletePayrollItem(ctx context.Context, req *entity.DeletePayrollItemReq) error
	ExportPayrollItemsPeriodically(ctx context.Context, req *entity.ExportPayrollItemsPeriodicallyReq) (*entity.ExportPayrollItemsPeriodicallyResp, error)
	GetPayrollReportsYearly(ctx context.Context, req *entity.GetPayrollReportsYearlyReq) (*entity.PayrollReportsYearlyResp, error)
	ResetPayrollRunItems(ctx context.Context, req *entity.ResetPayrollRunItemsReq) error
}
