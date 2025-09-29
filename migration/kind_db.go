package migration

type KindDB int

// type db
const (
	SqlServer KindDB = iota
	Postgres
	MySQL
)

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
