package core

import (
	"testing"

	"github.com/russross/meddler"
)

type moneyTest struct {
	Id   int64    `meddler:"id,pk"`
	Cash Monetary `meddler:"cash,monetary"`
}

func TestMonetary0DP(t *testing.T) {
	doMoneyTest(t, "AUD", "62", 36, 6200)
}

func TestMonetary1DP(t *testing.T) {
	doMoneyTest(t, "AUD", "62.69", 36, 6269)
}

func TestMonetaryErrorFormat(t *testing.T) {
	doMoneyTest(t, "AUD", "62.69.34", 0, 0)
}

func TestMonetaryUnknownCurrency(t *testing.T) {
	doMoneyTest(t, "BOOBOODOLLAR", "62.69", 0, 0)
}

func doMoneyTest(t *testing.T, curr string, amt string, expectedIso int16, expectedAmount int64) {
	c := NewTestConfig(t)
	ctx, err := NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	_, err = ctx.DB.Exec("CREATE TEMPORARY TABLE money_test (id BIGSERIAL PRIMARY KEY, cash monetary)")
	if err != nil {
		t.Fatal(err)
	}

	cash, err := NewMonetary(ctx.DB, curr, amt)
	if err != nil {
		if expectedIso == 0 && expectedAmount == 0 {
			return
		}
		t.Fatal(err)
	}

	m := moneyTest{0, cash}

	err = meddler.Insert(ctx.DB, "money_test", &m)
	if err != nil {
		t.Fatal(err)
	}

	m2 := moneyTest{}
	err = meddler.Load(ctx.DB, "money_test", &m2, m.Id)
	if err != nil {
		t.Fatal(err)
	}

	if m2.Cash.Iso4217Code != expectedIso {
		t.Fatal("ISO code incorrect")
	}

	if m2.Cash.Amount != expectedAmount {
		t.Fatal("Amount incorrect")
	}
}
