package log

// ANSI color definitions used for colored output.
const (
	ResetColor = "\033[0m"
	Green      = "\033[32m" // For prompts/questions and success messages.
	Blue       = "\033[34m" // For YAML field names and "go up" text.
	Purple     = "\033[35m" // For enum option values.
	White      = "\033[37m" // For schema descriptions and "generic" white text.
	Yellow     = "\033[33m" // For save option (always printed as 's').
	Red        = "\033[31m" // For error messages.
	Cyan       = "\033[36m" // Used for the full "Enter choice" prompt.
)
