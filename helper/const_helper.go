package helper

// datetime
const (
	YYYY_MM_DD       = "2006-01-02"
	DD_MM_YYYY       = "02-01-2006"
	DD_MM_YYYY_SLASH = "02-01-2006"
	DD_MMM_YYYY      = "02-Jan-2006"
	YYYYMMDDHHMMSS   = "20060102150405"

	// Date time ISO 8601
	DATETIME_FULL  = "2006-01-02T15:04:05Z07:00"
	DATETIME_NO_TZ = "2006-01-02 15:04:05.000"

	// format time
	TIME_FULL   = "150405"
	TIME_NO_SEC = "1504"

	// string time
	Second = "Second"
	Minute = "Minute"
	Hour   = "Hour"
	Day    = "Day"
	Month  = "Month"
	Year   = "Year"
)

// header
const (
	ContentType         = "Content-Type"
	Authorization       = "Authorization"
	ApplicationJSON     = "application/json"
	ApplicationFormData = "application/x-www-form-urlencoded"
)

// form data
const (
	ClientId          = "client_id"
	ClientSecret      = "client_secret"
	GrantType         = "grant_type"
	ClientCredentials = "client_credentials"
)

// type db
const (
	SQLServer = "SQLServer"
	Oracle    = "Oracle"
	Postgres  = "Postgres"
)
