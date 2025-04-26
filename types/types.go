package types

import (
	"github.com/shopspring/decimal"
)

// Money type
type Money = decimal.Decimal
type VND = int64

// KindDB type
type KindDB int

const (
	SqlServer KindDB = iota
	Postgres
	Oracle
	MySQL
)

func (k KindDB) String() string {
	switch k {
	case SqlServer:
		return "sqlserver"
	case Postgres:
		return "postgres"
	case Oracle:
		return "oracle"
	case MySQL:
		return "mysql"
	default:
		return ""
	}
}

func (k KindDB) GetDriver() string {
	switch k {
	case SqlServer:
		return "sqlserver"
	case Postgres:
		return "postgres"
	case Oracle:
		return "godror"
	case MySQL:
		return "mysql"
	default:
		return ""
	}
}

func (k KindDB) GetDialect() string {
	switch k {
	case SqlServer:
		return "mssql"
	case Postgres:
		return "postgres"
	case MySQL:
		return "mysql"
	default:
		return ""
	}
}
