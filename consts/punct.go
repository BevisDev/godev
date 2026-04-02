package consts

// punctuation — quotes & apostrophe
const (
	Apostrophe          = "'"
	QuoteSingle         = "'"
	QuoteDouble         = `"`
	QuoteBacktick       = "`"
	QuoteLeftGuillemet  = "«"
	QuoteRightGuillemet = "»"
)

// punctuation — brackets & grouping
const (
	ParenLeft    = "("
	ParenRight   = ")"
	BracketLeft  = "["
	BracketRight = "]"
	BraceLeft    = "{"
	BraceRight   = "}"
	AngleLeft    = "<"
	AngleRight   = ">"
)

// punctuation — sentence & clause
const (
	Comma       = ","
	Semicolon   = ";"
	Colon       = ":"
	Period      = "."
	Ellipsis    = "…"
	Exclamation = "!"
	Question    = "?"
	Interpunct  = "·"
)

// punctuation — dash & hyphen variants
const (
	Hyphen     = "-" // ASCII hyphen-minus
	MinusSign  = "\u2212"
	EnDash     = "\u2013"
	EmDash     = "\u2014"
	FigureDash = "\u2012"
)

// punctuation — separators & slashes
const (
	SlashForward = "/"
	SlashBack    = `\`
	Pipe         = "|"
	Ampersand    = "&"
	Asterisk     = "*"
	Percent      = "%"
	At           = "@"
	Hash         = "#"
	Dollar       = "$"
	Caret        = "^"
	Tilde        = "~"
	Equals       = "="
	Plus         = "+"
	Underscore   = "_"
)

// punctuation — space variants (for joining/splitting text)
const (
	Empty   = ""
	Space   = " "
	Tab     = "\t"
	Newline = "\n"
	CR      = "\r"
	NBSP    = "\u00a0" // no-break space
)
