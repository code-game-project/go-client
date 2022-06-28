package cg

import (
	"fmt"

	"github.com/mattn/go-colorable"
)

type color string

const (
	Reset  color = "\x1b[0m"
	Red    color = "\x1b[31m"
	Green  color = "\x1b[32m"
	Yellow color = "\x1b[33m"
)

var out = colorable.NewColorableStdout()

func printColor(color color, format string, a ...any) {
	fmt.Fprintf(out, "%s%s%s\n", color, fmt.Sprintf(format, a...), Reset)
}

func printWarning(format string, a ...any) {
	printColor(Yellow, fmt.Sprintf("WARNING: "+format, a...))
}

func printError(format string, a ...any) error {
	printColor(Red, "ERROR: "+format, a...)
	return fmt.Errorf(format, a...)
}
