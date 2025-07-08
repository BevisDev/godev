package types

// KindDB type
type KindDB int

type DBJSONTemplate string

const (
	TemplateJSONArray  DBJSONTemplate = "array"
	TemplateJSONObject DBJSONTemplate = "object"
)

// template
const (
	// MSSQLJSONArrayTemplate is a SQL Server template that wraps a SELECT query
	// in `FOR JSON PATH`, returning a JSON array.
	// If the result is NULL, it returns an empty JSON array (`[]`).
	MSSQLJSONArrayTemplate = `
	SELECT ISNULL((
		%s
		FOR JSON PATH
	), '[]') as data
	`

	// MSSQLJSONObjectTemplate is a SQL Server template that wraps a SELECT query
	// with `FOR JSON PATH, WITHOUT_ARRAY_WRAPPER` to return a single JSON object.
	MSSQLJSONObjectTemplate = `
	%s
	FOR JSON PATH, WITHOUT_ARRAY_WRAPPER;
	`

	// MySQLJSONArrayTemplate is a MySQL template for generating a JSON array
	// from rows using JSON_ARRAYAGG and JSON_OBJECT.
	MySQLJSONArrayTemplate = `
	SELECT IFNULL(JSON_ARRAYAGG(
	JSON_OBJECT(
		%s
	)), JSON_ARRAY()) AS data
	FROM %s
	%s
	`

	// MySQLJSONObjectTemplate returns a single JSON object
	MySQLJSONObjectTemplate = `
	SELECT JSON_OBJECT(
		%s
	) AS data
	FROM %s
	%s
	`

	// PostgresJSONArrayTemplate is a PostgreSQL SQL template that wraps a subquery
	// to return a JSON array using json_agg and row_to_json.
	PostgresJSONArrayTemplate = `
	SELECT COALESCE(
		json_agg(row_to_json(t)),
		'[]'::json
	) AS data
	FROM (
		%s
	) AS t;
	`

	// PostgresJSONObjectTemplate is a PostgreSQL SQL template that wraps a subquery
	// to return a single JSON object using row_to_json.
	PostgresJSONObjectTemplate = `
	SELECT row_to_json(t) AS data
	FROM (
		%s
	) AS t;
	`
)

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
