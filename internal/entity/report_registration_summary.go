package entity

import (
	"codebase-app/pkg/errmsg"
	"time"

	"github.com/shopspring/decimal"
)

type GetSummariesReq struct {
	UserID string `json:"user_id" validate:"ulid"`

	PaidAtFrom string `query:"paid_at_from" validate:"datetime=2006-01-02"`
	PaidAtTo   string `query:"paid_at_to" validate:"datetime=2006-01-02"`
	Timezone   string `query:"timezone" validate:"timezone"`
}

func (r *GetSummariesReq) SetDefault() {
	// Set default timezone first so it can be used for date calculations
	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
	}

	// Load the location for the timezone
	loc, err := time.LoadLocation(r.Timezone)
	if err != nil {
		// Fallback to system timezone if the specified timezone is invalid
		loc = time.Local
	}

	// Use the timezone for consistent date operations
	now := time.Now().In(loc)

	if r.PaidAtFrom == "" {
		// First day of current month in the specified timezone
		r.PaidAtFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Format("2006-01-02")
	}

	if r.PaidAtTo == "" {
		// Current date in the specified timezone
		r.PaidAtTo = now.Format("2006-01-02")
	}
}

func (r *GetSummariesReq) Validate() error {
	err := errmsg.NewCustomErrors(400)

	if (r.PaidAtFrom != "" || r.PaidAtTo != "") && (r.PaidAtFrom == "" || r.PaidAtTo == "") {
		err.Add("paid_at_from", "batas bawah tanggal pembayaran harus diisi")
		err.Add("paid_at_to", "batas atas tanggal pembayaran harus diisi")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

type GetSummariesResp struct {
	PaidAtFrom string `json:"paid_at_from"`
	PaidAtTo   string `json:"paid_at_to"`

	TotalHrFee               decimal.Decimal `json:"total_hr_fee"`
	TotalOverpaymentFee      decimal.Decimal `json:"total_overpayment_fee"`
	TotalMarketerCommission  decimal.Decimal `json:"total_marketer_commission_fee"`
	TotalMarketerGifts       decimal.Decimal `json:"total_marketer_gifts_fee"`
	TotalClosingFeeForReward decimal.Decimal `json:"total_closing_fee_for_reward"`
	TotalClosingFeeForOffice decimal.Decimal `json:"total_closing_fee_for_office"`
	TotalIncome              decimal.Decimal `json:"total_income"`
	TotalProfit              decimal.Decimal `json:"total_profit"`
}

type GetSummariesForCFO2Resp struct {
	PaidAtFrom string `json:"paid_at_from"`
	PaidAtTo   string `json:"paid_at_to"`

	TotalDebit       decimal.Decimal `json:"total_debit"`
	TotalOverpayment decimal.Decimal `json:"total_overpayment"`
	TotalCredit      decimal.Decimal `json:"total_credit"`
	TotalBalance     decimal.Decimal `json:"total_balance"`
}
