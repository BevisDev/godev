package migration

type DBType int

// type db
const (
	SqlServer DBType = iota + 1
	Postgres
	MySQL
)

func (d DBType) GetDialect() string {
	switch d {
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
