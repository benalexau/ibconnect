package gateway

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/benalexau/ibconnect/core"
	"github.com/gofinance/ib"
	"github.com/gorhill/cronexpr"
	"github.com/russross/meddler"
)

type AccountFeedFactory struct {
	AccountRefresh *cronexpr.Expression
}

func (f *AccountFeedFactory) NewFeed(ctx *FeedContext) *Feed {
	a := &AccountFeed{}
	notifications := []core.NtType{core.NtRefreshAll, core.NtAccountRefresh}
	callback := a.callback
	a.generic = NewGenericFeed(ctx, f.AccountRefresh, notifications, callback)
	var feed Feed = a
	return &feed
}

func (f *AccountFeedFactory) Done() core.NtType {
	return core.NtAccountFeedDone
}

type AccountFeed struct {
	generic   *GenericFeed
	tx        *sql.Tx                                         // scope is single callback only
	fc        *FeedContext                                    // scope is single callback only
	pam       *ib.PrimaryAccountManager                       // scope is single callback only
	created   time.Time                                       // scope is single callback only
	snapshots map[core.Account]core.AccountSnapshot           // scope is single callback only
	amounts   map[core.AccountSnapshot]core.AccountAmount     // scope is single callback only
	positions map[core.AccountSnapshot][]core.AccountPosition // scope is single callback only
}

func (a *AccountFeed) Close() {
	a.generic.Close()
}

func (a *AccountFeed) callback(ctx *FeedContext) {
	pam, err := ib.NewPrimaryAccountManager(ctx.Eng)
	if err != nil {
		ctx.Errors <- FeedError{err, a}
		return
	}

	defer pam.Close()
	var m ib.Manager = pam
	_, err = ib.SinkManager(m, 60*time.Second, 1)
	if err != nil {
		ctx.Errors <- FeedError{err, a}
		return
	}

	a.fc = ctx
	a.pam = pam
	a.created = time.Now()
	a.snapshots = make(map[core.Account]core.AccountSnapshot)
	a.amounts = make(map[core.AccountSnapshot]core.AccountAmount)
	a.positions = make(map[core.AccountSnapshot][]core.AccountPosition)

	defer func() {
		a.tx = nil
		a.fc = nil
		a.pam = nil
		a.created = time.Time{}
		a.snapshots = nil
		a.amounts = nil
		a.positions = nil
	}()

	err = a.processResults()
	if err != nil {
		ctx.Errors <- FeedError{err, a}
		return
	}

}

// processResults inserts into the database in a single transaction.
func (a *AccountFeed) processResults() error {
	var err error
	a.tx, err = a.fc.DB.Begin()
	if err != nil {
		return fmt.Errorf("gateway: account_feed begin TX: %v", err)
	}

	err = a.amount()
	if err != nil {
		a.tx.Rollback()
		return fmt.Errorf("gateway: account_feed amount: %v", err)
		return err
	}

	err = a.position()
	if err != nil {
		a.tx.Rollback()
		return fmt.Errorf("gateway: account_feed position: %v", err)
		return err
	}

	err = a.store()
	if err != nil {
		a.tx.Rollback()
		return fmt.Errorf("gateway: account_feed store: %v", err)
		return err
	}

	err = a.tx.Commit()
	if err != nil {
		return fmt.Errorf("gateway: account_feed commit TX: %v", err)
		return err
	}

	a.fc.N.Publish(core.NtAccountFeedDone, 1)
	return nil
}

func (a *AccountFeed) amount() error {
	for key, value := range a.pam.Values() {
		snapshot, err := a.getSnapshot(key.AccountCode)
		if err != nil {
			return fmt.Errorf("get snapshot %v", err)
		}

		amt := a.amounts[snapshot]
		amt.AccountSnapshotId = snapshot.Id

		if value.Currency == "BASE" {
			continue
		}

		switch key.Key {
		case "AccountType":
			val, err := a.getAccountType(value.Value)
			if err != nil {
				return fmt.Errorf("account type %v", err)
			}
			amt.AccountType = val.Id
		case "Cushion":
			val, err := strconv.ParseFloat(value.Value, 64)
			if err != nil {
				return err
			}
			amt.Cushion = val
		case "LookAheadNextChange":
			val, err := strconv.Atoi(value.Value)
			if err != nil {
				return err
			}
			amt.LookAheadNextChange = int16(val)
		case "AccruedCash":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("AccruedCash %s %s %v", value.Currency, value.Value, err)
			}
			amt.AccruedCash = val
		case "AvailableFunds":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("AvailableFunds %s %s %v", value.Currency, value.Value, err)
			}
			amt.AvailableFunds = val
		case "BuyingPower":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("BuyingPower %s %s %v", value.Currency, value.Value, err)
			}
			amt.BuyingPower = val
		case "EquityWithLoanValue":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("EquityWithLoanValue %s %s %v", value.Currency, value.Value, err)
			}
			amt.EquityWithLoanValue = val
		case "ExcessLiquidity":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("ExcessLiquidity %s %s %v", value.Currency, value.Value, err)
			}
			amt.ExcessLiquidity = val
		case "FullAvailableFunds":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("FullAvailableFunds %s %s %v", value.Currency, value.Value, err)
			}
			amt.FullAvailableFunds = val
		case "FullExcessLiquidity":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("FullExcessLiquidity %s %s %v", value.Currency, value.Value, err)
			}
			amt.FullExcessLiquidity = val
		case "FullInitMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("FullInitMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.FullInitMarginReq = val
		case "FullMaintMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("FullMaintMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.FullMaintMarginReq = val
		case "GrossPositionValue":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("GrossPositionValue %s %s %v", value.Currency, value.Value, err)
			}
			amt.GrossPositionValue = val
		case "InitMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("InitMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.InitMarginReq = val
		case "LookAheadAvailableFunds":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("LookAheadAvailableFunds %s %s %v", value.Currency, value.Value, err)
			}
			amt.LookAheadAvailableFunds = val
		case "LookAheadExcessLiquidity":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("LookAheadExcessLiquidity %s %s %v", value.Currency, value.Value, err)
			}
			amt.LookAheadExcessLiquidity = val
		case "LookAheadInitMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("LookAheadInitMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.LookAheadInitMarginReq = val
		case "LookAheadMaintMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("LookAheadMaintMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.LookAheadMaintMarginReq = val
		case "MaintMarginReq":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("MaintMarginReq %s %s %v", value.Currency, value.Value, err)
			}
			amt.MaintMarginReq = val
		case "NetLiquidation":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("NetLiquidation %s %s %v", value.Currency, value.Value, err)
			}
			amt.NetLiquidation = val
		case "TotalCashBalance":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("TotalCashBalance %s %s %v", value.Currency, value.Value, err)
			}
			amt.TotalCashValue = val
		case "TotalCashValue":
			val, err := core.NewMonetary(a.tx, value.Currency, value.Value)
			if err != nil {
				return fmt.Errorf("TotalCashValue %s %s %v", value.Currency, value.Value, err)
			}
			amt.TotalCashValue = val
		default:
			continue
		}
		a.amounts[snapshot] = amt
	}
	return nil
}

func (a *AccountFeed) position() error {
	for key, value := range a.pam.Portfolio() {
		snapshot, err := a.getSnapshot(key.AccountCode)
		if err != nil {
			return err
		}

		newPosition := new(core.AccountPosition)
		newPosition.AccountSnapshotId = snapshot.Id

		c := new(core.Contract)
		c.IbContractId = value.Contract.ContractID

		iso := new(core.Iso4217)
		err = meddler.QueryRow(a.tx, iso, "SELECT * FROM iso_4217 WHERE alphabetic_code = $1", value.Contract.Currency)
		if err != nil {
			return err
		}
		c.Iso4217Code = iso.Iso4217Code

		symbol, err := a.getSymbol(value.Contract.Symbol)
		if err != nil {
			return err
		}
		c.SymbolId = symbol.Id

		localSymbol, err := a.getSymbol(value.Contract.LocalSymbol)
		if err != nil {
			return err
		}
		c.LocalSymbolId = localSymbol.Id

		secType, err := a.getSecurityType(value.Contract.SecurityType)
		if err != nil {
			return err
		}
		c.SecurityTypeId = secType.Id

		exg, err := a.getExchange(value.Contract.PrimaryExchange)
		if err != nil {
			return err
		}
		c.PrimaryExchangeId = exg.Id

		con, err := a.getContract(*c)
		if err != nil {
			return err
		}
		newPosition.ContractId = con.Id

		newPosition.Position = value.Position
		newPosition.MarketPrice = value.MarketPrice
		newPosition.MarketValue = value.MarketValue
		newPosition.AverageCost = value.AverageCost
		newPosition.UnrealizedPNL = value.UnrealizedPNL
		newPosition.RealizedPNL = value.RealizedPNL

		knownPositions := a.positions[snapshot]
		knownPositions = append(knownPositions, *newPosition)
		a.positions[snapshot] = knownPositions
	}
	return nil
}

// getSnapshot returns the correct snapshot to use for this account key,
// taking care to create the records when required.
func (a *AccountFeed) getSnapshot(accountKey string) (core.AccountSnapshot, error) {
	acct, err := a.getAccount(accountKey)
	if err != nil {
		return core.AccountSnapshot{}, err
	}

	existing, ok := a.snapshots[acct]
	if ok {
		return existing, nil
	}

	snapshot, err := a.createAccountSnapshot(acct.Id)
	if err != nil {
		return core.AccountSnapshot{}, err
	}
	a.snapshots[acct] = snapshot

	return snapshot, nil
}

// getAccount returns the Account object, creating a database record if needed.
func (a *AccountFeed) getAccount(accountKey string) (core.Account, error) {
	existing := new(core.Account)
	err := meddler.QueryRow(a.tx, existing, "SELECT * FROM account WHERE account_code = $1", accountKey)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	acct := &core.Account{}
	acct.AccountCode = accountKey
	err = meddler.Insert(a.tx, "account", acct)
	return *acct, err
}

// createAccountSnapshot creates an AccountSnapshot object.
func (a *AccountFeed) createAccountSnapshot(accountId int64) (core.AccountSnapshot, error) {
	snap := &core.AccountSnapshot{}
	snap.AccountId = accountId
	snap.Created = a.created
	err := meddler.Insert(a.tx, "account_snapshot", snap)
	return *snap, err
}

// getAccountType returns the AccountType object, creating a database record if needed.
func (a *AccountFeed) getAccountType(desc string) (core.AccountType, error) {
	existing := new(core.AccountType)
	err := meddler.QueryRow(a.tx, existing, "SELECT * FROM account_type WHERE type_desc = $1", desc)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	at := &core.AccountType{}
	at.TypeDescription = desc
	err = meddler.Insert(a.tx, "account_type", at)
	return *at, err
}

// getSecurityType returns the SecurityType object, creating a database record if needed.
func (a *AccountFeed) getSecurityType(desc string) (core.SecurityType, error) {
	existing := new(core.SecurityType)
	err := meddler.QueryRow(a.tx, existing, "SELECT * FROM security_type WHERE security_type = $1", desc)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	st := &core.SecurityType{}
	st.SecurityType = desc
	err = meddler.Insert(a.tx, "security_type", st)
	return *st, err
}

// getSymbol returns the Symbol object, creating a database record if needed.
func (a *AccountFeed) getSymbol(desc string) (core.Symbol, error) {
	existing := new(core.Symbol)
	err := meddler.QueryRow(a.tx, existing, "SELECT * FROM symbol WHERE symbol = $1", desc)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	s := &core.Symbol{}
	s.Symbol = desc
	err = meddler.Insert(a.tx, "symbol", s)
	return *s, err
}

// getExchange returns the Exchange object, creating a database record if needed.
func (a *AccountFeed) getExchange(desc string) (core.Exchange, error) {
	existing := new(core.Exchange)
	err := meddler.QueryRow(a.tx, existing, "SELECT * FROM exchange WHERE exchange = $1", desc)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	e := &core.Exchange{}
	e.Exchange = desc
	err = meddler.Insert(a.tx, "exchange", e)
	return *e, err
}

// getContract returns the Contract object, creating a database record if needed.
func (a *AccountFeed) getContract(criteria core.Contract) (core.Contract, error) {
	existing := new(core.Contract)
	err := meddler.QueryRow(a.tx, existing,
		"SELECT * FROM contract WHERE ib_contract_id = $1 AND "+
			"iso_4217_code = $2 AND symbol_id = $3 AND local_symbol_id = $4 AND "+
			"security_type_id = $5 AND primary_exchange_id = $6",
		criteria.IbContractId, criteria.Iso4217Code, criteria.SymbolId,
		criteria.LocalSymbolId, criteria.SecurityTypeId, criteria.PrimaryExchangeId)
	if err != nil && err != sql.ErrNoRows {
		return *existing, err
	}

	if existing.Id != 0 {
		return *existing, nil
	}

	criteria.Created = a.created
	err = meddler.Insert(a.tx, "contract", &criteria)
	return criteria, err
}

// store writes the full updates into the database in a single transaction.
func (a *AccountFeed) store() error {
	for _, amt := range a.amounts {
		err := meddler.Insert(a.tx, "account_amount", &amt)
		if err != nil {
			return err
		}
	}

	for _, pos := range a.positions {
		for _, p := range pos {
			err := meddler.Insert(a.tx, "account_position", &p)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
