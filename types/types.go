package types

import "github.com/shopspring/decimal"

// Money type
type Money = decimal.Decimal
type VND = int64

// KindDB type
type KindDB string

func (k KindDB) String() string {
	return string(k)
}

const (
	SqlServer KindDB = "SQLServer"
	Postgres  KindDB = "Postgres"
	Oracle    KindDB = "Oracle"
	MySQL     KindDB = "MySQL"
)

// DriverDB type
type DriverDB string

func (d DriverDB) String() string {
	return string(d)
}

const (
	SqlServerDriver DriverDB = "sqlserver"
	PostgresDriver  DriverDB = "postgres"
	GodrorDriver    DriverDB = "godror"
	MySQLDriver     DriverDB = "mysql"
)

// SQLDriver mapping
var SQLDriver = map[KindDB]DriverDB{
	SqlServer: SqlServerDriver,
	Postgres:  PostgresDriver,
	Oracle:    GodrorDriver,
	MySQL:     MySQLDriver,
}

// Dialect type
type Dialect string

func (d Dialect) String() string {
	return string(d)
}

const (
	MSSQLDialect    Dialect = "mssql"
	PostgresDialect Dialect = "postgres"
	MySQLDialect    Dialect = "mysql"
)

// DialectMigration mapping
var DialectMigration = map[KindDB]Dialect{
	SqlServer: MSSQLDialect,
	Postgres:  PostgresDialect,
	MySQL:     MySQLDialect,
}
