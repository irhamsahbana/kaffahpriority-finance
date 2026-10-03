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
	return buildFixtureWithHeaders(t, fixtureHeaders, rows)
}

// buildFixtureWithHeaders is buildFixture with an explicit header map, for
// fixtures whose layout differs from the current-format template.
func buildFixtureWithHeaders(t *testing.T, headers map[string]string, rows []map[string]any) []byte {
	t.Helper()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := f.SetSheetName("Sheet1", fixtureSheet); err != nil {
		t.Fatal(err)
	}
	for col, header := range headers {
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

// legacyHeaders mirrors the pre-feature-fee template: no PC-LN headers,
// KEEP GAJI at P, PAYROLL ITEM ID at S.
var legacyHeaders = map[string]string{
	"F": "JUMLAH", "G": "HITUNGAN", "H": "UJROH FULL", "J": "TF/F",
	"K": "FL", "L": "NL", "P": "KEEP GAJI", "S": "PAYROLL ITEM ID",
}

// legacyRow builds one data row for the fallback layout. An empty id string
// writes an empty ID cell at S; a nil keepWage leaves P unwritten.
func legacyRow(id string, keepWage any) map[string]any {
	row := map[string]any{
		"F": 10, "G": 1000.0, "H": 1000.0, "J": "F",
		"K": 500.0, "L": 500.0, "S": id,
	}
	if keepWage != nil {
		row["P"] = keepWage
	}
	return row
}

// TestImportPayrollItems_LegacyLayout_NoFeeHeaders covers a file with none of
// the six fee headers: the fallback layout resolves IDs to S and KEEP GAJI to
// P, and every feature fee routes through the empty-raw branch to zero.
func TestImportPayrollItems_LegacyLayout_NoFeeHeaders(t *testing.T) {
	id := ulid.Make().String()
	xlsx := buildFixtureWithHeaders(t, legacyHeaders, []map[string]any{legacyRow(id, 25000.0)})
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

	for _, fc := range featureCols {
		assertFee(t, fc.get(items[0]), 0, fc.name)
	}
	assertDecimal(t, items[0].ForeignLearningFee, 500, "FL")
	assertDecimal(t, items[0].NightLearningFee, 500, "NL")
	if items[0].Wage.Equal(decimal.NewFromInt(25000)) == false {
		t.Fatalf("KEEP GAJI: expected 25000, got %s", items[0].Wage)
	}
	if items[0].ID != id {
		t.Fatalf("ID: expected %s, got %s", id, items[0].ID)
	}
}

// TestImportPayrollItems_PartialFeeHeaders covers MT-LN headers at N1-R1 with
// no PC header: hasFeatureColumns stays false, so Wage reads the fallback
// column P (which doubles as the MS fee column) and IDs come from S.
func TestImportPayrollItems_PartialFeeHeaders(t *testing.T) {
	id := ulid.Make().String()
	headers := map[string]string{
		"F": "JUMLAH", "G": "HITUNGAN", "H": "UJROH FULL", "J": "TF/F",
		"N": "MT", "O": "CL", "P": "MS", "Q": "SC", "R": "LN",
	}
	xlsx := buildFixtureWithHeaders(t, headers, []map[string]any{{
		"F": 10, "G": 1000.0, "H": 1000.0, "J": "F",
		"N": 1001.0, "O": 1002.0, "P": 1003.0, "Q": 1004.0, "R": 1005.0,
		"S": id,
	}})
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

	// PC header absent -> empty raw -> zero; MT-LN captured from their headers.
	assertFee(t, items[0].PCFee, 0, "PC")
	assertFee(t, items[0].MTFee, 1001, "MT")
	assertFee(t, items[0].CLFee, 1002, "CL")
	assertFee(t, items[0].MSFee, 1003, "MS")
	assertFee(t, items[0].SCFee, 1004, "SC")
	assertFee(t, items[0].LNFee, 1005, "LN")
	// Wage falls back to column P, which this fixture populates with the MS
	// sentinel instead of 25000.
	if items[0].Wage.Equal(decimal.NewFromInt(1003)) == false {
		t.Fatalf("Wage: expected 1003, got %s", items[0].Wage)
	}
	// FL/NL fall back to empty columns K/L.
	assertDecimal(t, items[0].ForeignLearningFee, 0, "FL")
	assertDecimal(t, items[0].NightLearningFee, 0, "NL")
	if items[0].ID != id {
		t.Fatalf("ID: expected %s, got %s", id, items[0].ID)
	}
}

// TestImportPayrollItems_NearMissHeader_PCFEE covers a "PC FEE" header, which
// is not EqualFold "PC": hasFeatureColumns stays false, so PC reads no column
// and KEEP GAJI / Payroll Item ID come from the fallback columns P and S.
func TestImportPayrollItems_NearMissHeader_PCFEE(t *testing.T) {
	id := ulid.Make().String()
	headers := map[string]string{
		"F": "JUMLAH", "G": "HITUNGAN", "H": "UJROH FULL", "J": "TF/F",
		"M": "PC FEE", "P": "KEEP GAJI", "S": "PAYROLL ITEM ID",
	}
	xlsx := buildFixtureWithHeaders(t, headers, []map[string]any{{
		"F": 10, "G": 1000.0, "H": 1000.0, "J": "F",
		"P": 25000.0, "S": id,
	}})
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

	for _, fc := range featureCols {
		assertFee(t, fc.get(items[0]), 0, fc.name)
	}
	if items[0].Wage.Equal(decimal.NewFromInt(25000)) == false {
		t.Fatalf("Wage: expected 25000, got %s", items[0].Wage)
	}
	if items[0].ID != id {
		t.Fatalf("ID: expected %s, got %s", id, items[0].ID)
	}
}

// TestImportPayrollItems_LegacyLayout_EmptyKeepWageCell is characterization:
// an empty KEEP GAJI cell at P yields Wage 0 and nil AcquisitionRights both
// before and after the empty-fee-cell fix.
func TestImportPayrollItems_LegacyLayout_EmptyKeepWageCell(t *testing.T) {
	id := ulid.Make().String()
	xlsx := buildFixtureWithHeaders(t, legacyHeaders, []map[string]any{legacyRow(id, "")})
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

	if items[0].Wage.Equal(decimal.Zero) == false {
		t.Fatalf("Wage: expected 0, got %s", items[0].Wage)
	}
	if items[0].AcquisitionRights != nil {
		t.Fatalf("AcquisitionRights: expected nil, got %d", *items[0].AcquisitionRights)
	}
}

// TestImportPayrollItems_LegacyLayout_EmptyIDCell_SkipsRow is characterization:
// a row with an empty Payroll Item ID cell is silently skipped before any
// parsing, so only the valid row reaches the repository.
func TestImportPayrollItems_LegacyLayout_EmptyIDCell_SkipsRow(t *testing.T) {
	validID := ulid.Make().String()
	xlsx := buildFixtureWithHeaders(t, legacyHeaders, []map[string]any{
		legacyRow("", 25000.0),
		legacyRow(validID, 25000.0),
	})
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
	if items[0].ID != validID {
		t.Fatalf("ID: expected %s, got %s", validID, items[0].ID)
	}
}
