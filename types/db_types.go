package types

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

func (k KindDB) GetPlaceHolder() string {
	switch k {
	case SqlServer:
		return "@p"
	case Postgres:
		return "$"
	default: // mysql
		return "?"
	}
}

func (k KindDB) GetConnectionString() string {
	switch k {
	case SqlServer:
		return "sqlserver://%s:%s@%s:%d?database=%s"
	case Postgres:
		return "postgres://%s:%s@%s:%d/%s?sslmode=disable"
	case Oracle:
		return "%s/%s@%s:%d/%s"
	case MySQL:
		return "%s:%s@tcp(%s:%d)/%s"
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
