package database

var connectionMap = map[DBType]string{
	// username/password/host/port/db
	SqlServer: "sqlserver://%s:%s@%s:%d?database=%s",

	// username/password/host/port/db
	Postgres: "postgres://%s:%s@%s:%d/%s?sslmode=disable",

	// username/password@host:port/db
	Oracle: "%s/%s@%s:%d/%s",

	// username/password/host/port/db
	MySQL: "%s:%s@tcp(%s:%d)/%s",
}
