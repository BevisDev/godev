package consts

// pattern
const (
	Email         = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	TenDigitPhone = `^\d{10}$`
	UUID          = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`
	AlphaNumeric  = `^[a-zA-Z0-9]+$`
	DateYYYYMMDD  = `^\d{4}-\d{2}-\d{2}$`
	IPv4          = `^(\d{1,3}\.){3}\d{1,3}$`
	VNIDNumber    = `^\d{9}(\d{3})?$`
	FilePattern   = `^[\w,\s-]+\.[A-Za-z0-9]{1,8}$`
)
