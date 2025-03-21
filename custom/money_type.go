package custom

import (
	"github.com/BevisDev/godev/helper"
	"github.com/shopspring/decimal"
)

type Money decimal.Decimal

func (m *Money) UnmarshalJSON(b []byte) error {
	var d decimal.Decimal
	if err := helper.ToStruct(b, &d); err != nil {
		return err
	}
	*m = Money(d)
	return nil
}
