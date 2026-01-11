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
