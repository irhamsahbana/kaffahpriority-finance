package entity

import (
	"codebase-app/pkg/errmsg"
	"time"

	"github.com/shopspring/decimal"
)

type GetSummariesReq struct {
	UserId string `json:"user_id" validate:"ulid"`

	PaidAtFrom string `query:"paid_at_from" validate:"datetime=2006-01-02"`
	PaidAtTo   string `query:"paid_at_to" validate:"datetime=2006-01-02"`
	Timezone   string `query:"timezone" validate:"timezone"`
}

func (r *GetSummariesReq) SetDefault() {
	if r.PaidAtFrom == "" {
		r.PaidAtFrom = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")
	}

	if r.PaidAtTo == "" {
		r.PaidAtTo = time.Now().Format("2006-01-02")
	}

	if r.Timezone == "" {
		r.Timezone = "Asia/Makassar"
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
	TotalProfit              decimal.Decimal `json:"total_profit"`
}
