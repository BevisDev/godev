package consts

// header
const (
	ContentType                  = "Content-Type"
	ContentDisposition           = "Content-Disposition"
	ContentDispositionInline     = `inline; filename="%s"`
	ContentDispositionAttachment = `attachment; filename="%s"`

	Authorization          = "Authorization"
	ApplicationJSON        = "application/json"
	ApplicationFormData    = "application/x-www-form-urlencoded"
	ApplicationOctetStream = "application/octet-stream"

	// ApplicationPDF for using PDF
	ApplicationPDF = "application/pdf"

	// ApplicationMSWord office 97 - 2003 (.doc, .xls, .ppt)
	ApplicationMSWord       = "application/msword"
	ApplicationMSExcel      = "application/vnd.ms-excel"
	ApplicationMSPowerPoint = "application/vnd.ms-powerpoint"

	// ApplicationWordOpenXML (.docx, .xlsx, .pptx)
	ApplicationWordOpenXML  = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ApplicationExcelOpenXML = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ApplicationPptOpenXML   = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// ApplicationZip (.zip, .7z, windows zip)
	ApplicationZip  = "application/zip"
	ApplicationX7z  = "application/x-7z-compressed"
	ApplicationXZip = "application/x-zip-compressed"

	// ApplicationXML for using xml
	ApplicationXML = "application/xml"
	TextXML        = "text/xml"

	// TextCSV ( commons text)
	TextCSV           = "text/csv"
	TextPlain         = "text/plain"
	TextHTML          = "text/html"
	MultipartFormData = "multipart/form-data"

	// ImagePNG for using image
	ImagePNG  = "image/png"
	ImageJPEG = "image/jpeg"
	ImageGIF  = "image/gif"

	// VideoMP4 for using video
	VideoMP4  = "video/mp4"
	VideoMPEG = "video/mpeg"
)

// auth
const (
	Bearer_ = "Bearer "
	Basic_  = "Basic "
)

// form data
const (
	ClientId          = "client_id"
	ClientSecret      = "client_secret"
	GrantType         = "grant_type"
	ClientCredentials = "client_credentials"
	AuthorizationCode = "authorization_code"
	Username          = "username"
	Password          = "password"
	AccessToken       = "access_token"
	RefreshToken      = "refresh_token"
	Scope             = "scope"
)

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

const (
	State    = "state"
	Status   = "status"
	Header   = "header"
	Body     = "body"
	Duration = "duration"
	Time     = "time"
	Method   = "method"
	Url      = "url"
	Query    = "query"
)

// extension
const (
	// Text files
	ExtTXT  = "txt"
	ExtCSV  = "csv"
	ExtJSON = "json"
	ExtXML  = "xml"
	ExtYAML = "yaml"
	ExtMD   = "md"

	// Images
	ExtJPG  = "jpg"
	ExtJPEG = "jpeg"
	ExtPNG  = "png"
	ExtGIF  = "gif"
	ExtBMP  = "bmp"
	ExtSVG  = "svg"
	ExtWEBP = "webp"

	// Archives
	ExtZIP = "zip"
	ExtRAR = "rar"
	Ext7Z  = "7z"
	ExtTAR = "tar"
	ExtGZ  = "gz"

	// Documents
	ExtPDF  = "pdf"
	ExtDOC  = "doc"
	ExtDOCX = "docx"
	ExtXLS  = "xls"
	ExtXLSX = "xlsx"
	ExtPPT  = "ppt"
	ExtPPTX = "pptx"

	// Code
	ExtGO   = "go"
	ExtJS   = "js"
	ExtTS   = "ts"
	ExtHTML = "html"
	ExtCSS  = "css"
	ExtSQL  = "sql"
	ExtJAVA = "java"
	ExtPY   = "py"

	// Video & Audio
	ExtMP4 = "mp4"
	ExtAVI = "avi"
	ExtMKV = "mkv"
	ExtMOV = "mov"
	ExtMP3 = "mp3"
	ExtWAV = "wav"
)
