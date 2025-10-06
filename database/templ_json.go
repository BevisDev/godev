package database

// TemplateJSON type to get json
type TemplateJSON int

const (
	TemplateJSONArray TemplateJSON = iota
	TemplateJSONObject
)

var TemplateDBMap = map[DBType]map[TemplateJSON]string{
	SqlServer: {
		TemplateJSONArray:  MSSQLJSONArrayTemplate,
		TemplateJSONObject: MSSQLJSONObjectTemplate,
	},
	Postgres: {
		TemplateJSONArray:  PostgresJSONArrayTemplate,
		TemplateJSONObject: PostgresJSONObjectTemplate,
	},
	MySQL: {
		TemplateJSONArray:  MySQLJSONArrayTemplate,
		TemplateJSONObject: MySQLJSONObjectTemplate,
	},
}

// MSSQL templates
// MSSQLJSONArrayTemplate returns a JSON array, If the result is NULL, it returns an empty JSON array (`[]`).
// MSSQLJSONObjectTemplate returns a single JSON object.
const (
	MSSQLJSONArrayTemplate = `
	SELECT ISNULL((
		%s
		FOR JSON PATH
	), '[]') as data
	`

	MSSQLJSONObjectTemplate = `
	%s
	FOR JSON PATH, WITHOUT_ARRAY_WRAPPER;
	`
)

// Postgres templates
// PostgresJSONArrayTemplate returns a JSON array using json_agg and row_to_json.
// PostgresJSONObjectTemplate returns a single JSON object using row_to_json.
const (
	PostgresJSONArrayTemplate = `
	SELECT COALESCE(
		json_agg(row_to_json(t)),
		'[]'::json
	) AS data
	FROM (
		%s
	) AS t;
	`

	PostgresJSONObjectTemplate = `
	SELECT row_to_json(t) AS data
	FROM (
		%s
	) AS t;
	`
)

// MySQL templates are split and require manual composition.
// In MySQL, the JSON templates require explicit table and WHERE clause placeholders.
// You must use fmt.Sprintf(template, columns, table, where) when applying this.
// MySQLJSONArrayTemplate returns a JSON array
// MySQLJSONObjectTemplate returns a single JSON object
const (
	MySQLJSONArrayTemplate = `
	SELECT IFNULL(JSON_ARRAYAGG(
	JSON_OBJECT(
		%s
	)), JSON_ARRAY()) AS data
	FROM %s
	%s
	`

	MySQLJSONObjectTemplate = `
	SELECT JSON_OBJECT(
		%s
	) AS data
	FROM %s
	%s
	`
)
