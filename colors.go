package main

import (
	"fmt"
	"runtime"
)

const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"

	ColorRed     = "\033[91m"
	ColorGreen   = "\033[92m"
	ColorYellow  = "\033[93m"
	ColorBlue    = "\033[94m"
	ColorMagenta = "\033[95m"
	ColorCyan    = "\033[96m"
	ColorWhite   = "\033[97m"
)

func init() {
	if runtime.GOOS == "windows" {
		fmt.Print("\033[?25h")
	}
}

func colorPrint(color, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s%s\n", color, msg, ColorReset)
}

func colorPrintInline(color, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s%s", color, msg, ColorReset)
}

func printSuccess(format string, a ...interface{}) {
	colorPrint(ColorGreen, format, a...)
}

func printError(format string, a ...interface{}) {
	colorPrint(ColorRed, format, a...)
}

func printInfo(format string, a ...interface{}) {
	colorPrint(ColorCyan, format, a...)
}

func printWarning(format string, a ...interface{}) {
	colorPrint(ColorYellow, format, a...)
}

func printHeader(format string, a ...interface{}) {
	colorPrint(ColorBold+ColorMagenta, format, a...)
}

func printStep(format string, a ...interface{}) {
	colorPrintInline(ColorWhite, "  ")
	colorPrintInline(ColorCyan, "» ")
	colorPrintInline(ColorWhite, format, a...)
	fmt.Println()
}

func printStepOK(hostName, url string) {
	colorPrintInline(ColorWhite, "  ")
	colorPrintInline(ColorGreen, "✓ ")
	colorPrintInline(ColorBold+ColorWhite, "%s", hostName)
	colorPrintInline(ColorDim, " → ")
	fmt.Println(url)
}

func printStepError(hostName, errMsg string) {
	colorPrintInline(ColorWhite, "  ")
	colorPrintInline(ColorRed, "✗ ")
	colorPrintInline(ColorBold+ColorWhite, "%s", hostName)
	colorPrintInline(ColorDim, " : ")
	colorPrintln(ColorRed, errMsg)
}

func colorPrintln(color, msg string) {
	fmt.Printf("%s%s%s\n", color, msg, ColorReset)
}
