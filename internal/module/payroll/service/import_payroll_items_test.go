package service

import (
	"bytes"
	"context"
	"math"
	"strconv"
	"testing"

	"codebase-app/internal/entity"
	ports "codebase-app/internal/ports/module/payroll"
	"codebase-app/pkg/errmsg"

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

// TestImportPayrollItems_PCColumnNumericEdgeMatrix pins the fee parser's
// grammar: cells parse via decimal.NewFromString then InexactFloat64 (not
// strconv.ParseFloat). Rejection cases must surface exactly one message under
// key row_2 and never reach the repository.
//
// Each case runs against its own isolated fixture containing exactly one data
// row at sheet row 2, so the 1e400 case's captured +Inf pointer cannot collide
// with any other case and every rejection provably aborts the whole import.
func TestImportPayrollItems_PCColumnNumericEdgeMatrix(t *testing.T) {
	cases := []struct {
		name     string
		cellVal  string
		ok       bool
		fee      float64 // ok && !overflow: expected PCFee value
		overflow bool    // ok && PCFee is +Inf (1e400 infrastructure path)
		errMsg   string  // !ok: the exact single message under key row_2
	}{
		{name: "scientific_notation_1e3_stores_1000", cellVal: "1e3", ok: true, fee: 1000},
		{name: "leading_plus_sign_5_stores_5", cellVal: "+5", ok: true, fee: 5},
		{name: "exponent_overflow_1e400_stores_plus_inf", cellVal: "1e400", ok: true, overflow: true},
		{name: "reject_nan", cellVal: "NaN", errMsg: "PC tidak valid: NaN"},
		{name: "reject_inf", cellVal: "Inf", errMsg: "PC tidak valid: Inf"},
		{name: "reject_plus_inf", cellVal: "+Inf", errMsg: "PC tidak valid: +Inf"},
		{name: "reject_minus_inf", cellVal: "-Inf", errMsg: "PC tidak valid: -Inf"},
		{name: "reject_hex_float", cellVal: "0x1p3", errMsg: "PC tidak valid: 0x1p3"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xlsx := buildFixture(t, []map[string]any{sentinelRow("M", tc.cellVal)})
			repo := &capturingRepo{}
			svc := NewPayrollService(Config{Repo: repo})
			resp, err := svc.ImportPayrollItems(context.Background(), &entity.ImportPayrollItemsReq{
				File:     bytes.NewReader(xlsx),
				FileName: "rekap-gaji.xlsx",
			})

			if tc.ok {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if resp.TotalProcessed != 1 || resp.TotalUpdated != 1 {
					t.Fatalf("expected success response {1, 1}, got {processed: %d, updated: %d}", resp.TotalProcessed, resp.TotalUpdated)
				}
				if repo.importCalls != 1 {
					t.Fatalf("expected 1 repo call, got %d", repo.importCalls)
				}
				items := repo.importArgs[0].Items
				if len(items) != 1 {
					t.Fatalf("expected 1 item, got %d", len(items))
				}
				fee := items[0].PCFee
				if fee == nil {
					t.Fatal("PCFee: expected non-nil *float64, got nil")
				}
				if tc.overflow {
					if !math.IsInf(*fee, 1) {
						t.Fatalf("PCFee: expected +Inf, got %v", *fee)
					}
				} else if *fee != tc.fee {
					t.Fatalf("PCFee: expected %v, got %v", tc.fee, *fee)
				}
				// Guard against column mix-ups: MT sentinel must stay intact.
				assertFee(t, items[0].MTFee, 1001, "MT")
				return
			}

			if err == nil {
				t.Fatalf("expected rejection with %q, got success", tc.errMsg)
			}
			if repo.importCalls != 0 {
				t.Fatalf("expected 0 repo calls, got %d", repo.importCalls)
			}
			cerr, ok := err.(*errmsg.CustomError)
			if !ok {
				t.Fatalf("expected *errmsg.CustomError, got %T", err)
			}
			msgs := cerr.Errors["row_2"]
			if len(msgs) != 1 || msgs[0] != tc.errMsg {
				t.Fatalf("expected Errors[row_2] = [%q], got %#v", tc.errMsg, msgs)
			}
		})
	}
}
