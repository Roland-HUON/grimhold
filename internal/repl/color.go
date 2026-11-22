package repl

import (
	"fmt"
	"strings"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBold    = "\033[1m"
	colorMagenta = "\033[35m"
)

func colorize(text, color string) string {
	return color + text + colorReset
}

func hpBar(hp, maxHP, width int) string {
	var color string
	if maxHP <= 0 {
		return ""
	}
	filled := hp * width / maxHP
	if filled < 0 {
		filled = 0
	} else if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percent := hp * 100 / maxHP
	if percent > 50 {
		color = colorGreen
	} else if percent > 25 {
		color = colorYellow
	} else {
		color = colorRed
	}
	return colorize("["+bar+"]", color) + fmt.Sprintf(" %d/%d HP", hp, maxHP)
}

func dangerColor(danger int) string {
	switch danger {
	case 1:
		return colorGreen
	case 2:
		return colorYellow
	case 3:
		return colorRed
	default:
		return colorBold + colorMagenta
	}
}
