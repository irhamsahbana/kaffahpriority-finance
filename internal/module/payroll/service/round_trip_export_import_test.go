package service

import (
	"context"
	"os"
	"testing"

	"codebase-app/internal/entity"
	ports "codebase-app/internal/ports/module/payroll"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

// exportStubRepo is a hand-written test double implementing only the three
// repo methods the export service calls (export_payroll_items.go). It embeds
// ports.PayrollRepository so it satisfies the interface; any other method
// panics on the nil embedded interface, which is fine because the export flow
// never calls them.
type exportStubRepo struct {
	ports.PayrollRepository

	run   *entity.PayrollRun
	items []entity.PayrollItem
}

func (r *exportStubRepo) GetPayrollRunByPeriod(_ context.Context, _ string) (*entity.PayrollRun, error) {
	return r.run, nil
}

func (r *exportStubRepo) GetPayrollItemsForExport(_ context.Context, _ string) ([]entity.PayrollItem, error) {
	return r.items, nil
}

// GetAdditionalStudents returns an empty, non-nil map so the exported sheet's
// data rows are exactly the fixture items.
func (r *exportStubRepo) GetAdditionalStudents(_ context.Context, _ []string) (map[string][]entity.PayrollItemAdditionalStudent, error) {
	return map[string][]entity.PayrollItemAdditionalStudent{}, nil
}

// roundTripFixtureItems builds items covering both sides of every fee branch:
// feature fees are either nil (export omits the cell) or non-zero (export
// writes it); a zero *float64 is deliberately absent because
// export_payroll_items.go:266 skips it, making it indistinguishable from nil
// in the file. FL/NL likewise mix non-zero and zero.
func roundTripFixtureItems() []entity.PayrollItem {
	pcFee := 1500.0
	scFee := 2250.5

	full := entity.PayrollItem{
		ID:                  ulid.Make().String(),
		AcademicManagerName: "AM Satu",
		LecturerName:        "Ustadz A",
		StudentName:         "Santri X",
		ProgramName:         "Program P",
		MarketerName:        "Marketer M",
		ProgramMeetings:     8,
		IsMeetingFull:       true,
		WagePerMeeting:      decimal.NewFromInt(1000),
		FullWage:            decimal.NewFromInt(8000),
		InitialWage:         decimal.NewFromInt(8000),
		ForeignLearningFee:  decimal.NewFromInt(500),
		NightLearningFee:    decimal.NewFromInt(750),
		FeatureFees: entity.FeatureFees{
			PCFee: &pcFee,
			SCFee: &scFee,
		},
		Wage:              decimal.NewFromInt(25000),
		AcquisitionRights: 3,
	}

	// Zero-value feature fees / FL / NL / Wage on purpose: the export omits
	// those cells and the import must map them back to zero, not nil.
	partial := entity.PayrollItem{
		ID:                  ulid.Make().String(),
		AcademicManagerName: "AM Satu",
		LecturerName:        "Ustadz B",
		StudentName:         "Santri Y",
		ProgramName:         "Program Q",
		MarketerName:        "Marketer N",
		ProgramMeetings:     6,
		IsMeetingFull:       false,
		WagePerMeeting:      decimal.NewFromInt(500),
		FullWage:            decimal.NewFromInt(3000),
		InitialWage:         decimal.NewFromInt(3000),
	}

	return []entity.PayrollItem{full, partial}
}

// payrollItemFee reads a feature fee off the export fixture by spreadsheet
// column (M-R), mirroring featureCols.
func payrollItemFee(item entity.PayrollItem, col string) *float64 {
	switch col {
	case "M":
		return item.PCFee
	case "N":
		return item.MTFee
	case "O":
		return item.CLFee
	case "P":
		return item.MSFee
	case "Q":
		return item.SCFee
	case "R":
		return item.LNFee
	}
	return nil
}

// importedItemFee reads a feature fee off the captured import item by
// spreadsheet column (M-R), mirroring featureCols.
func importedItemFee(item entity.ImportedPayrollItems, col string) *float64 {
	switch col {
	case "M":
		return item.PCFee
	case "N":
		return item.MTFee
	case "O":
		return item.CLFee
	case "P":
		return item.MSFee
	case "Q":
		return item.SCFee
	case "R":
		return item.LNFee
	}
	return nil
}

// expectedInitialWage mirrors the recomputation in import_payroll_items.go
// (:205-209): full meetings take FullWage as-is, otherwise WagePerMeeting is
// multiplied by the meetings parsed from column F.
func expectedInitialWage(item entity.PayrollItem) decimal.Decimal {
	if item.IsMeetingFull {
		return item.FullWage
	}
	return item.WagePerMeeting.Mul(decimal.NewFromInt(int64(item.ProgramMeetings)))
}

func assertDecimalEqual(t *testing.T, got decimal.Decimal, want decimal.Decimal, label string) {
	t.Helper()
	if got.Equal(want) == false {
		t.Fatalf("%s: expected %s, got %s", label, want, got)
	}
}

func assertFeatureDecimal(t *testing.T, got *decimal.Decimal, want decimal.Decimal, label string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: expected non-nil *decimal.Decimal, got nil", label)
	}
	assertDecimalEqual(t, *got, want, label)
}

func TestRoundTripExportImport(t *testing.T) {
	items := roundTripFixtureItems()
	fixtureByID := make(map[string]entity.PayrollItem, len(items))
	for _, it := range items {
		fixtureByID[it.ID] = it
	}

	exportSvc := NewPayrollService(Config{Repo: &exportStubRepo{
		run:   &entity.PayrollRun{ID: ulid.Make().String()},
		items: items,
	}})
	exportResp, err := exportSvc.ExportPayrollItemsPeriodically(context.Background(), &entity.ExportPayrollItemsPeriodicallyReq{Period: "2026-10"})
	if err != nil {
		t.Fatalf("ExportPayrollItemsPeriodically returned error: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(exportResp.FilePath) })

	// Feed the file unmodified: read the bytes the export service wrote to
	// disk, never re-saving or editing them.
	xlsx, err := os.ReadFile(exportResp.FilePath)
	if err != nil {
		t.Fatalf("read exported file %s: %v", exportResp.FilePath, err)
	}

	repo, importResp := runImport(t, xlsx)

	if repo.importCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.importCalls)
	}
	if importResp.TotalProcessed != len(items) {
		t.Fatalf("expected TotalProcessed %d, got %d", len(items), importResp.TotalProcessed)
	}

	captured := repo.importArgs[0].Items
	if len(captured) != len(items) {
		t.Fatalf("expected %d captured items, got %d", len(items), len(captured))
	}

	for _, got := range captured {
		want, ok := fixtureByID[got.ID]
		if !ok {
			t.Fatalf("captured item ID %q not found in fixture", got.ID)
		}

		if got.ProgramMeetings != int(want.ProgramMeetings) {
			t.Fatalf("item %s ProgramMeetings: expected %d, got %d", got.ID, want.ProgramMeetings, got.ProgramMeetings)
		}
		if got.IsMeetingFull != want.IsMeetingFull {
			t.Fatalf("item %s IsMeetingFull: expected %v, got %v", got.ID, want.IsMeetingFull, got.IsMeetingFull)
		}
		assertDecimalEqual(t, got.FullWage, want.FullWage, "FullWage")
		assertDecimalEqual(t, got.Wage, want.Wage, "Wage")

		for _, fc := range featureCols {
			wantPtr := payrollItemFee(want, fc.col)
			if wantPtr == nil {
				// Omitted (nil/zero) fee re-imports as a non-nil pointer to 0.
				assertFee(t, importedItemFee(got, fc.col), 0, fc.name)
			} else {
				assertFee(t, importedItemFee(got, fc.col), *wantPtr, fc.name)
			}
		}

		assertFeatureDecimal(t, got.ForeignLearningFee, want.ForeignLearningFee, "FL")
		assertFeatureDecimal(t, got.NightLearningFee, want.NightLearningFee, "NL")

		// InitialWage is recomputed on import (column I is ignored), so the
		// expectation is the recomputation from F/G/H/J, not a read-back.
		assertDecimalEqual(t, got.InitialWage, expectedInitialWage(want), "InitialWage")
	}
}
