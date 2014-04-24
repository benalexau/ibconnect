package core

import "time"

type AccountType struct {
	Id              int64  `meddler:"id,pk"`
	TypeDescription string `meddler:"type_desc"`
}

type Account struct {
	Id          int64  `meddler:"id,pk" json:"-"`
	AccountCode string `meddler:"account_code"`
}

type AccountSnapshot struct {
	Id        int64     `meddler:"id,pk"`
	AccountId int64     `meddler:"account_id"`
	Created   time.Time `meddler:"created,utctime"`
}

type AccountSnapshotLatest struct {
	AccountCode string    `meddler:"account_code"`
	Latest      time.Time `meddler:"latest,utctime"`
}

type AccountAmount struct {
	Id                       int64    `meddler:"id,pk"`
	AccountSnapshotId        int64    `meddler:"account_snapshot_id"`
	AccountType              int64    `meddler:"account_type_id"`
	Cushion                  float64  `meddler:"cushion"`
	LookAheadNextChange      int16    `meddler:"look_ahead_next_change"`
	AccruedCash              Monetary `meddler:"accrued_cash,monetary"`
	AvailableFunds           Monetary `meddler:"available_funds,monetary"`
	BuyingPower              Monetary `meddler:"buying_power,monetary"`
	EquityWithLoanValue      Monetary `meddler:"excess_liquidity,monetary"`
	ExcessLiquidity          Monetary `meddler:"equity_with_loan_value,monetary"`
	FullAvailableFunds       Monetary `meddler:"full_available_funds,monetary"`
	FullExcessLiquidity      Monetary `meddler:"full_excess_liquidity,monetary"`
	FullInitMarginReq        Monetary `meddler:"full_init_margin_req,monetary"`
	FullMaintMarginReq       Monetary `meddler:"full_maint_margin_req,monetary"`
	GrossPositionValue       Monetary `meddler:"gross_position_value,monetary"`
	InitMarginReq            Monetary `meddler:"init_margin_req,monetary"`
	LookAheadAvailableFunds  Monetary `meddler:"look_ahead_available_funds,monetary"`
	LookAheadExcessLiquidity Monetary `meddler:"look_ahead_excess_liquidity,monetary"`
	LookAheadInitMarginReq   Monetary `meddler:"look_ahead_init_margin_req,monetary"`
	LookAheadMaintMarginReq  Monetary `meddler:"look_ahead_maint_margin_req,monetary"`
	MaintMarginReq           Monetary `meddler:"maint_margin_req,monetary"`
	NetLiquidation           Monetary `meddler:"net_liquidation,monetary"`
	TotalCashBalance         Monetary `meddler:"total_cash_balance,monetary"`
	TotalCashValue           Monetary `meddler:"total_cash_value,monetary"`
}

type AccountAmountView struct {
	AccountSnapshotId        int64   `meddler:"account_snapshot_id,pk" json:"-"`
	AccountType              string  `meddler:"account_type"`
	Cushion                  float64 `meddler:"cushion"`
	LookAheadNextChange      int16   `meddler:"look_ahead_next_change"`
	AccruedCash              string  `meddler:"accrued_cash"`
	AvailableFunds           string  `meddler:"available_funds"`
	BuyingPower              string  `meddler:"buying_power"`
	EquityWithLoanValue      string  `meddler:"excess_liquidity"`
	ExcessLiquidity          string  `meddler:"equity_with_loan_value"`
	FullAvailableFunds       string  `meddler:"full_available_funds"`
	FullExcessLiquidity      string  `meddler:"full_excess_liquidity"`
	FullInitMarginReq        string  `meddler:"full_init_margin_req"`
	FullMaintMarginReq       string  `meddler:"full_maint_margin_req"`
	GrossPositionValue       string  `meddler:"gross_position_value"`
	InitMarginReq            string  `meddler:"init_margin_req"`
	LookAheadAvailableFunds  string  `meddler:"look_ahead_available_funds"`
	LookAheadExcessLiquidity string  `meddler:"look_ahead_excess_liquidity"`
	LookAheadInitMarginReq   string  `meddler:"look_ahead_init_margin_req"`
	LookAheadMaintMarginReq  string  `meddler:"look_ahead_maint_margin_req"`
	MaintMarginReq           string  `meddler:"maint_margin_req"`
	NetLiquidation           string  `meddler:"net_liquidation"`
	TotalCashBalance         string  `meddler:"total_cash_balance"`
	TotalCashValue           string  `meddler:"total_cash_value"`
}

type SecurityType struct {
	Id           int64  `meddler:"id,pk"`
	SecurityType string `meddler:"security_type"`
}

type Symbol struct {
	Id     int64  `meddler:"id,pk"`
	Symbol string `meddler:"symbol"`
}

type Exchange struct {
	Id       int64  `meddler:"id,pk"`
	Exchange string `meddler:"exchange"`
}

type Contract struct {
	Id                int64     `meddler:"id,pk"`
	Created           time.Time `meddler:"created,utctime"`
	IbContractId      int64     `meddler:"ib_contract_id"`
	Iso4217Code       int16     `meddler:"iso_4217_code"`
	SymbolId          int64     `meddler:"symbol_id"`
	LocalSymbolId     int64     `meddler:"local_symbol_id"`
	SecurityTypeId    int64     `meddler:"security_type_id"`
	PrimaryExchangeId int64     `meddler:"primary_exchange_id"`
}

type AccountPosition struct {
	Id                int64   `meddler:"id,pk"`
	AccountSnapshotId int64   `meddler:"account_snapshot_id"`
	ContractId        int64   `meddler:"contract_id"`
	Position          int64   `meddler:"pos"`
	MarketPrice       float64 `meddler:"market_price"`
	MarketValue       float64 `meddler:"market_value"`
	AverageCost       float64 `meddler:"average_cost"`
	UnrealizedPNL     float64 `meddler:"unrealized_pnl"`
	RealizedPNL       float64 `meddler:"realized_pnl"`
}

type AccountPositionView struct {
	IbContractId      int64     `meddler:"ib_contract_id,pk"`
	Symbol            string    `meddler:"symbol"`
	LocalSymbol       string    `meddler:"local_symbol"`
	SecurityType      string    `meddler:"security_type"`
	Exchange          string    `meddler:"exchange"`
	Position          int64     `meddler:"pos"`
	Iso4217Code       int16     `meddler:"iso_4217_code"`
	Currency          string    `meddler:"currency"`
	MarketPrice       float64   `meddler:"market_price"`
	MarketValue       float64   `meddler:"market_value"`
	AverageCost       float64   `meddler:"average_cost"`
	UnrealizedPNL     float64   `meddler:"unrealized_pnl"`
	RealizedPNL       float64   `meddler:"realized_pnl"`
	AccountSnapshotId int64     `meddler:"account_snapshot_id" json:"-"`
	Created           time.Time `meddler:"created,utctime" json:"-"`
	AccountCode       string    `meddler:"account_code" json:"-"`
}
