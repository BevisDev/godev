package database

type DBType int

// type db
const (
	SqlServer DBType = iota + 1
	Postgres
	Oracle
	MySQL
)

func (d DBType) String() string {
	switch d {
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

func (d DBType) ConnectionString() string {
	switch d {
	case SqlServer:
		// username/password/host/port/db
		return "sqlserver://%s:%s@%s:%d?database=%s"
	case Postgres:
		// username/password/host/port/db
		return "postgres://%s:%s@%s:%d/%s?sslmode=disable"
	case Oracle:
		// username/password@host:port/db
		return "%s/%s@%s:%d/%s"
	case MySQL:
		// username/password/host/port/db
		return "%s:%s@tcp(%s:%d)/%s"
	default:
		return ""
	}
}

func (d DBType) GetDriver() string {
	switch d {
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

func (d DBType) GetPlaceHolder() string {
	switch d {
	case SqlServer:
		return "@p"
	case Postgres:
		return "$"
	default: // mysql
		return "?"
	}
}
