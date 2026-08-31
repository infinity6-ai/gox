package logzcolor

import "fmt"

type Color string

const RESET = "\033[0m"

const (
	// Regular Colors
	BLACK   Color = "\033[30m"
	RED     Color = "\033[31m"
	GREEN   Color = "\033[32m"
	YELLOW  Color = "\033[33m"
	BLUE    Color = "\033[34m"
	MAGENTA Color = "\033[35m"
	CYAN    Color = "\033[36m"
	WHITE   Color = "\033[37m"

	// Bold Colors
	BOLD_RED     Color = "\033[1;31m"
	BOLD_GREEN   Color = "\033[1;32m"
	BOLD_YELLOW  Color = "\033[1;33m"
	BOLD_BLUE    Color = "\033[1;34m"
	BOLD_MAGENTA Color = "\033[1;35m"
	BOLD_CYAN    Color = "\033[1;36m"
	BOLD_WHITE   Color = "\033[1;37m"

	// Bright Colors
	BRIGHT_BLACK   Color = "\033[90m"
	BRIGHT_RED     Color = "\033[91m"
	BRIGHT_GREEN   Color = "\033[92m"
	BRIGHT_YELLOW  Color = "\033[93m"
	BRIGHT_BLUE    Color = "\033[94m"
	BRIGHT_MAGENTA Color = "\033[95m"
	BRIGHT_CYAN    Color = "\033[96m"
	BRIGHT_WHITE   Color = "\033[97m"

	// Bold Bright Colors
	BOLD_BRIGHT_BLACK   Color = "\033[1;90m"
	BOLD_BRIGHT_RED     Color = "\033[1;91m"
	BOLD_BRIGHT_GREEN   Color = "\033[1;92m"
	BOLD_BRIGHT_YELLOW  Color = "\033[1;93m"
	BOLD_BRIGHT_BLUE    Color = "\033[1;94m"
	BOLD_BRIGHT_MAGENTA Color = "\033[1;95m"
	BOLD_BRIGHT_CYAN    Color = "\033[1;96m"
	BOLD_BRIGHT_WHITE   Color = "\033[1;97m"
)

func (me Color) Apply(text string) string {
	return fmt.Sprintf("%s%s%s", me, text, RESET)
}
