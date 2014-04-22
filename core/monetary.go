package core

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/russross/meddler"
)

// Monetary represents a monetary amount in a specific currency. The amount is
// stored in minor units of the currency for reasons of space efficiency,
// formatting convenience and freedom from floating point precision issues.
type Monetary struct {
	Iso4217Code int16
	Amount      int64
}

// NewMonetary returns a money amount, associating it with the correct ISO 4217 record.
func NewMonetary(db meddler.DB, currency string, amount string) (Monetary, error) {
	m := new(Monetary)
	iso := new(Iso4217)
	err := meddler.QueryRow(db, iso, "SELECT * FROM iso_4217 WHERE alphabetic_code = $1", currency)
	if err != nil {
		return *m, err
	}
	m.Iso4217Code = iso.Iso4217Code

	var major, minor int

	split := strings.Split(amount, ".")
	if len(split) > 2 {
		return *m, fmt.Errorf("amount '%s' should be an integer or contain a single decimal point", amount)
	}

	if len(split) == 2 {
		major, err = strconv.Atoi(split[0])
		if err != nil {
			return *m, err
		}

		minor, err = strconv.Atoi(split[1])
		if err != nil {
			return *m, err
		}
	} else {
		major, err = strconv.Atoi(amount)
		if err != nil {
			return *m, err
		}
	}

	// convert it to cents based on what this currency uses
	m.Amount = (int64(major) * int64(math.Pow10(int(iso.MinorUnit)))) + int64(minor)
	return *m, nil
}

// Iso4217 represents officially-reported information about a specific currency.
type Iso4217 struct {
	Iso4217Code    int16  `meddler:"iso_4217_code"`
	MinorUnit      int16  `meddler:"minor_unit"`
	AlphabeticCode string `meddler:"alphabetic_code"`
	Currency       string `meddler:"currency"`
}

// MonetaryMeddler converts between Monetary values and the associated Postgres
// composite type.
type MonetaryMeddler struct{}

func (mm MonetaryMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	return new([]byte), nil
}

func (mm MonetaryMeddler) PostRead(fieldAddr, scanTarget interface{}) error {
	money, ok := fieldAddr.(*Monetary)
	if !ok {
		return fmt.Errorf("MonetaryMeddler.PostRead: not writing to a monetary field")
	}

	ptr := scanTarget.(*[]byte)
	if ptr == nil {
		return fmt.Errorf("MonetaryMeddler.PostRead: nil pointer")
	}
	raw := *ptr
	str := bytes.NewBuffer(raw).String()

	if !strings.HasPrefix(str, "(") || !strings.HasSuffix(str, ")") {
		return fmt.Errorf("MonetaryMeddler.PostRead: '%s' is not a composite type", str)
	}

	str = str[1 : len(str)-1] // drop ( and )
	split := strings.Split(str, ",")
	if len(split) != 2 {
		return fmt.Errorf("MonetaryMeddler.PostRead: '%s' did not have the expected 2 composite type column values", str)
	}

	iso, err := strconv.Atoi(split[0])
	if err != nil {
		return fmt.Errorf("MonetaryMeddler.PostRead: '%s' field '%s' is not an integer: %v", str, split[0], err)
	}
	money.Iso4217Code = int16(iso)

	amt, err := strconv.Atoi(split[1])
	if err != nil {
		return fmt.Errorf("MonetaryMeddler.PostRead: '%s' field '%s' is not an integer: %v", str, split[1], err)
	}
	money.Amount = int64(amt)

	return nil
}

func (mm MonetaryMeddler) PreWrite(field interface{}) (saveValue interface{}, err error) {
	buffer := new(bytes.Buffer)
	money, ok := field.(Monetary)
	if !ok {
		return nil, fmt.Errorf("MonetaryMeddler.PreWrite: not a monetary field")
	}
	fmt.Fprintf(buffer, "(%d, %d)", money.Iso4217Code, money.Amount)
	return buffer.Bytes(), nil
}
