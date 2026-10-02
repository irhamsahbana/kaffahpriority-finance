package entity

import "codebase-app/pkg/errmsg"

const UpgradeMultiplier = 2

type FeatureFees struct {
	PCFee *float64 `json:"pc_fee" db:"pc_fee" validate:"omitempty,min=0"`
	MTFee *float64 `json:"mt_fee" db:"mt_fee" validate:"omitempty,min=0"`
	CLFee *float64 `json:"cl_fee" db:"cl_fee" validate:"omitempty,min=0"`
	MSFee *float64 `json:"ms_fee" db:"ms_fee" validate:"omitempty,min=0"`
	SCFee *float64 `json:"sc_fee" db:"sc_fee" validate:"omitempty,min=0"`
	LNFee *float64 `json:"ln_fee" db:"ln_fee" validate:"omitempty,min=0"`
}

type FeatureFeeValues struct {
	PCFee float64
	MTFee float64
	CLFee float64
	MSFee float64
	SCFee float64
	LNFee float64
}

func ValidateUpgrades(isITP, isSSP bool) error {
	if isITP && isSSP {
		return errmsg.NewCustomErrors(400).SetMessage("ITP dan SSP tidak boleh aktif bersamaan")
	}
	return nil
}

func FeatureFeeValuesFromPointers(fees FeatureFees) FeatureFeeValues {
	valueOf := func(value *float64) float64 {
		if value == nil {
			return 0
		}
		return *value
	}
	return FeatureFeeValues{
		PCFee: valueOf(fees.PCFee),
		MTFee: valueOf(fees.MTFee),
		CLFee: valueOf(fees.CLFee),
		MSFee: valueOf(fees.MSFee),
		SCFee: valueOf(fees.SCFee),
		LNFee: valueOf(fees.LNFee),
	}
}

func TotalFeatureFee(fees FeatureFeeValues) float64 {
	return fees.PCFee + fees.MTFee + fees.CLFee + fees.MSFee + fees.SCFee + fees.LNFee
}
