package datetime

// Date & time layouts
const (
	// ---- Date only ----
	DateLayoutDMYDash  = "02-01-2006"
	DateLayoutDMYSlash = "02/01/2006"
	DateLayoutDMYMonth = "02-Jan-2006"
	DateLayoutCompact  = "20060102"
	DateLayoutISO      = "2006-01-02"

	// ---- DateTime (no timezone) ----
	DateTimeLayout      = "2006-01-02 15:04:05"
	DateTimeLayoutMilli = "2006-01-02 15:04:05.000"
	DateTimeLayoutLocal = "2006-01-02T15:04:05"

	// ---- DateTime with timezone ----
	DateTimeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
	DateTimeLayoutUTC     = "2006-01-02T15:04:05Z"

	// ---- Compact ----
	DateTimeLayoutCompact = "20060102150405"

	// format ISO 8601 / RFC3339
	DateTimeMillis = "2006-01-02 15:04:05.000"

	// ---- Time only ----
	TimeLayout          = "15:04:05"
	TimeLayoutCompact   = "150405"
	TimeLayoutNoSeconds = "1504"
)
