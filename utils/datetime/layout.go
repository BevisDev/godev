package datetime

// Layout datetime
const (
	// format common
	DD_MM_YYYYY       = "02-01-2006"
	DD_MM_YYYYY_FLASH = "02/01/2006"
	DD_MMM_YYYY       = "02-Jan-2006"
	YYYYMMDDHHMMSS    = "20060102150405"
	YYYYMMDD          = "20060102"

	// format ISO 8601 / RFC3339
	DateOnly       = "2006-01-02"
	DateTime       = "2006-01-02 15:04:05"
	DateTimeOffset = "2006-01-02T15:04:05Z07:00"
	DatetimeUTC    = "2006-01-02T15:04:05Z"
	DateTimeSQL    = "2006-01-02 15:04:05.000"
	DateTimeNoTZ   = "2006-01-02T15:04:05"

	// format time
	TimeOnly    = "15:04:05"
	TimeCompact = "150405"
	TimeNoSec   = "1504"
)
