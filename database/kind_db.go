package database

type KindDB int

// type db
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
	// go get github.com/denisenkom/go-mssqldb
	case SqlServer:
		return "sqlserver"

	// go get github.com/lib/pq
	case Postgres:
		return "postgres"

	// go get github.com/godror/godror
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
