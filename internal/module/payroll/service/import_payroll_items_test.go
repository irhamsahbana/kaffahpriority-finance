package service

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"codebase-app/internal/entity"
	ports "codebase-app/internal/ports/module/payroll"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// capturingRepo is a hand-written test double for ports.PayrollRepository.
// It embeds the interface and overrides only ImportPayrollItems; any other
// method call panics on the nil embedded interface, which is fine because the
// import flow never calls them.
type capturingRepo struct {
	ports.PayrollRepository

	importCalls int
	importArgs  []*entity.ImportPayrollItemsReq
}

var _ ports.PayrollRepository = (*capturingRepo)(nil)

func (r *capturingRepo) ImportPayrollItems(ctx context.Context, req *entity.ImportPayrollItemsReq) (int64, error) {
	r.importCalls++
	cp := *req
	// Copy the items slice so later mutations (if any) cannot alias the record.
	cp.Items = append([]entity.ImportedPayrollItems(nil), req.Items...)
	r.importArgs = append(r.importArgs, &cp)
	return int64(len(req.Items)), nil
}

const fixtureSheet = "Rekap Gaji"

// fixtureHeaders maps cell references in row 1 to header labels, mirroring the
// current-format template: PC-LN at M1-R1, FL S1, NL T1, KEEP GAJI V1,
// PAYROLL ITEM ID Y1.
var fixtureHeaders = map[string]string{
	"F": "JUMLAH", "G": "HITUNGAN", "H": "UJROH FULL", "J": "TF/F",
	"M": "PC", "N": "MT", "O": "CL", "P": "MS", "Q": "SC", "R": "LN",
	"S": "FL", "T": "NL", "V": "KEEP GAJI", "Y": "PAYROLL ITEM ID",
}

type featureCol struct {
	col      string
	name     string
	sentinel float64
	get      func(entity.ImportedPayrollItems) *float64
}

var featureCols = []featureCol{
	{"M", "PC", 1000, func(i entity.ImportedPayrollItems) *float64 { return i.PCFee }},
	{"N", "MT", 1001, func(i entity.ImportedPayrollItems) *float64 { return i.MTFee }},
	{"O", "CL", 1002, func(i entity.ImportedPayrollItems) *float64 { return i.CLFee }},
	{"P", "MS", 1003, func(i entity.ImportedPayrollItems) *float64 { return i.MSFee }},
	{"Q", "SC", 1004, func(i entity.ImportedPayrollItems) *float64 { return i.SCFee }},
	{"R", "LN", 1005, func(i entity.ImportedPayrollItems) *float64 { return i.LNFee }},
}

// buildFixture renders data rows (starting at row 2) into an in-memory xlsx
// with the current-format header layout. A nil cell value leaves the cell
// unwritten; any other value (including "") is written explicitly.
func buildFixture(t *testing.T, rows []map[string]any) []byte {
	t.Helper()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := f.SetSheetName("Sheet1", fixtureSheet); err != nil {
		t.Fatal(err)
	}
	for col, header := range fixtureHeaders {
		if err := f.SetCellValue(fixtureSheet, col+"1", header); err != nil {
			t.Fatal(err)
		}
	}
	for i, row := range rows {
		rowIdx := i + 2
		for col, val := range row {
			if err := f.SetCellValue(fixtureSheet, col+strconv.Itoa(rowIdx), val); err != nil {
				t.Fatal(err)
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// sentinelRow builds one row where every fee column carries its sentinel and
// only targetCol holds targetVal (nil = cell left unwritten). Non-fee required
// columns are always populated: JUMLAH 10, TF/F "F", wage cells 1000.
func sentinelRow(targetCol string, targetVal any) map[string]any {
	row := map[string]any{
		"F": 10, "G": 1000.0, "H": 1000.0, "J": "F",
		"S": 500.0, "T": 500.0, "V": 25000.0,
		"Y": ulid.Make().String(),
	}
	for _, fc := range featureCols {
		if fc.col == targetCol {
			if targetVal != nil {
				row[fc.col] = targetVal
			}
		} else {
			row[fc.col] = fc.sentinel
		}
	}
	return row
}

// runFeatureCellCase imports a fixture with a single sentinel-isolated row and
// asserts the target column maps to 0 while every sentinel stays unchanged.
func runFeatureCellCase(t *testing.T, fc featureCol, targetVal any) {
	t.Helper()

	xlsx := buildFixture(t, []map[string]any{sentinelRow(fc.col, targetVal)})
	repo, resp := runImport(t, xlsx)

	if repo.importCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.importCalls)
	}
	if resp.TotalProcessed != 1 {
		t.Fatalf("expected TotalProcessed 1, got %d", resp.TotalProcessed)
	}
	items := repo.importArgs[0].Items
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	for _, other := range featureCols {
		want := other.sentinel
		if other.col == fc.col {
			want = 0
		}
		assertFee(t, other.get(items[0]), want, other.name)
	}
	assertDecimal(t, items[0].ForeignLearningFee, 500, "FL")
	assertDecimal(t, items[0].NightLearningFee, 500, "NL")
	if items[0].Wage.Equal(decimal.NewFromInt(25000)) == false {
		t.Fatalf("KEEP GAJI: expected 25000, got %s", items[0].Wage)
	}
}

func runImport(t *testing.T, xlsx []byte) (*capturingRepo, *entity.ImportPayrollItemsResp) {
	t.Helper()

	repo := &capturingRepo{}
	svc := NewPayrollService(Config{Repo: repo})
	resp, err := svc.ImportPayrollItems(context.Background(), &entity.ImportPayrollItemsReq{
		File:     bytes.NewReader(xlsx),
		FileName: "rekap-gaji.xlsx",
	})
	if err != nil {
		t.Fatalf("ImportPayrollItems returned error: %v", err)
	}
	return repo, resp
}

func assertFee(t *testing.T, got *float64, want float64, label string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: expected non-nil *float64, got nil", label)
	}
	if *got != want {
		t.Fatalf("%s: expected %v, got %v", label, want, *got)
	}
}

func assertDecimal(t *testing.T, got *decimal.Decimal, want int64, label string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: expected non-nil *decimal.Decimal, got nil", label)
	}
	if got.Equal(decimal.NewFromInt(want)) == false {
		t.Fatalf("%s: expected %d, got %s", label, want, got)
	}
}

func TestImportPayrollItems_EmptyStringFeatureCell_MapsToZero(t *testing.T) {
	for _, fc := range featureCols {
		t.Run(fc.name, func(t *testing.T) {
			runFeatureCellCase(t, fc, "")
		})
	}
}

func TestImportPayrollItems_WhitespaceOnlyFeatureCell_MapsToZero(t *testing.T) {
	for _, fc := range featureCols {
		t.Run(fc.name+"_spaces", func(t *testing.T) {
			runFeatureCellCase(t, fc, "   ")
		})
	}
	// Collapse-pin: additional whitespace forms on PC column.
	for _, form := range []struct {
		label string
		val   string
	}{
		{"tab", "\t"},
		{"crlf", "\r\n"},
		{"nbsp", "\u00a0"},
	} {
		t.Run("PC_"+form.label, func(t *testing.T) {
			runFeatureCellCase(t, featureCols[0], form.val)
		})
	}
}

func TestImportPayrollItems_UnwrittenFeatureCell_MapsToZero(t *testing.T) {
	for _, fc := range featureCols {
		t.Run(fc.name, func(t *testing.T) {
			runFeatureCellCase(t, fc, nil)
		})
	}
}

// baseRow returns the non-fee columns every fixture row needs: JUMLAH 10,
// HITUNGAN 1000, UJROH FULL 1000, TF/F "F", KEEP GAJI 25000, and a valid ULID.
func baseRow() map[string]any {
	return map[string]any{
		"F": 10, "G": 1000.0, "H": 1000.0, "J": "F",
		"V": 25000.0, "Y": ulid.Make().String(),
	}
}

func TestImportPayrollItems_AllSixFeatureCellsEmpty_RowImportsAsZeroes(t *testing.T) {
	// Blank forms are exercised separately: unwritten (nil) mirrors the
	// export round-trip where zero fees are omitted, "" is an explicitly
	// cleared cell. FL/NL are left empty too, so this is the truly
	// all-empty row.
	for _, form := range []struct {
		label string
		val   any // nil = unwritten
	}{
		{"unwritten", nil},
		{"empty_string", ""},
	} {
		t.Run(form.label, func(t *testing.T) {
			row := baseRow()
			for _, fc := range featureCols {
				if form.val != nil {
					row[fc.col] = form.val
				}
			}

			xlsx := buildFixture(t, []map[string]any{row})
			repo, resp := runImport(t, xlsx) // runImport fails the test on any error

			if repo.importCalls != 1 {
				t.Fatalf("expected 1 repo call, got %d", repo.importCalls)
			}
			if resp.TotalProcessed != 1 {
				t.Fatalf("expected TotalProcessed 1, got %d", resp.TotalProcessed)
			}
			items := repo.importArgs[0].Items
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			// Captured-argument assertions stand in for DB-level NOT NULL
			// checks: the pq error string is deliberately never asserted.
			for _, fc := range featureCols {
				assertFee(t, fc.get(items[0]), 0, fc.name)
			}
			assertDecimal(t, items[0].ForeignLearningFee, 0, "FL")
			assertDecimal(t, items[0].NightLearningFee, 0, "NL")
		})
	}
}

func TestImportPayrollItems_MixedRow_EmptyCellsZeroKeptValuesPreserved(t *testing.T) {
	// PC unwritten, MT "", MS unwritten -> empty; CL/SC/LN non-empty.
	// SC carries an explicit written 0 so a parsed zero is distinguishable
	// from a blank-mapped zero only by route, both must land on &0.0.
	// FL and NL are left unwritten on purpose: their parse path is separate
	// from the six *float64 feature branches and must resolve to non-nil
	// *decimal.Decimal pointers to decimal.Zero.
	row := baseRow()
	row["O"] = 1234.5
	row["Q"] = 0
	row["R"] = 2500.75

	xlsx := buildFixture(t, []map[string]any{row})
	repo, resp := runImport(t, xlsx)

	if repo.importCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.importCalls)
	}
	if resp.TotalProcessed != 1 {
		t.Fatalf("expected TotalProcessed 1, got %d", resp.TotalProcessed)
	}
	items := repo.importArgs[0].Items
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	want := map[string]float64{"PC": 0, "MT": 0, "MS": 0, "CL": 1234.5, "SC": 0, "LN": 2500.75}
	for _, fc := range featureCols {
		assertFee(t, fc.get(items[0]), want[fc.name], fc.name)
	}
	assertDecimal(t, items[0].ForeignLearningFee, 0, "FL")
	assertDecimal(t, items[0].NightLearningFee, 0, "NL")
}
