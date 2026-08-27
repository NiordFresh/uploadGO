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

	EmojiOK      = "\u2714"
	EmojiError   = "\u2718"
	EmojiArrow   = "\u279C"
	EmojiUpload  = "\u2B07"
	EmojiLink    = "\u2197"
	EmojiFile    = "\u25A3"
	EmojiSuccess = "\u2714"
	EmojiFail    = "\u2718"
	EmojiPlay    = "\u25B6"
	EmojiStop    = "\u25A0"
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
	colorPrintInline(ColorCyan, EmojiPlay+" ")
	colorPrintInline(ColorWhite, format, a...)
	fmt.Println()
}

func printStepOK(hostName, url string) {
	colorPrintInline(ColorWhite, "  ")
	colorPrintInline(ColorGreen, EmojiOK+" ")
	colorPrintInline(ColorBold+ColorWhite, "%s", hostName)
	colorPrintInline(ColorDim, " "+EmojiArrow+" ")
	fmt.Println(url)
}

func printStepError(hostName, errMsg string) {
	colorPrintInline(ColorWhite, "  ")
	colorPrintInline(ColorRed, EmojiError+" ")
	colorPrintInline(ColorBold+ColorWhite, "%s", hostName)
	colorPrintInline(ColorDim, " : ")
	colorPrintln(ColorRed, errMsg)
}

func colorPrintln(color, msg string) {
	fmt.Printf("%s%s%s\n", color, msg, ColorReset)
}
