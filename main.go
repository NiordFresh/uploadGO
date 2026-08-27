package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"uploadGO/host"
)

func init() {
	if runtime.GOOS == "windows" {
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		proc := kernel32.NewProc("SetConsoleOutputCP")
		proc.Call(uintptr(65001))
	}
}

func main() {
	langFlag := flag.String("lang", "", "Language (pl or en)")
	installFlag := flag.Bool("install", false, "Install to Windows context menu")
	uninstallFlag := flag.Bool("uninstall", false, "Remove from Windows context menu")
	flag.Parse()

	if *installFlag {
		if err := GenerateRegFiles(); err != nil {
			printError("Error generating .reg files: %v", err)
			return
		}
		printSuccess("Reg files generated. Double-click uploadGO.reg to install.")
		return
	}

	if *uninstallFlag {
		printWarning("Double-click uploadGO_remove.reg to uninstall.")
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		cfg := LoadSettings()
		msg := GetMessages(cfg.Language)
		printHeader(msg.Usage)
		return
	}

	filePath := args[0]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		cfg := LoadSettings()
		msg := GetMessages(cfg.Language)
		printError("%s %s", msg.ErrorFileNotFound, filePath)
		return
	}

	cfg := LoadSettings()

	if *langFlag != "" {
		cfg.Language = *langFlag
	}

	msg := GetMessages(cfg.Language)

	enabledHosts := getEnabledHosts(cfg)
	if len(enabledHosts) == 0 {
		printError(msg.ErrorNoHost)
		return
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		printError("Error: %v", err)
		return
	}

	info, _ := os.Stat(absPath)
	fileSize := int64(0)
	if info != nil {
		fileSize = info.Size()
	}

	printHeader(msg.UploadingToAll)
	fmt.Println()

	var allResults []*host.UploadResult
	var allErrors []error

	for _, hostName := range enabledHosts {
		printStep("%s", hostName)

		progress := NewProgress(fileSize)

		result, err := uploadFile(hostName, absPath, progress, cfg)
		if err != nil {
			printStepError(hostName, err.Error())
			allErrors = append(allErrors, fmt.Errorf("%s: %w", hostName, err))
			continue
		}

		printStepOK(hostName, result.URL)
		allResults = append(allResults, result)
	}

	fmt.Println()
	printHeader(msg.Results)

	for i, r := range allResults {
		fmt.Println()
		colorPrint(ColorBold, "  %s %d. %s", EmojiFile, i+1, r.Filename)
		colorPrint(ColorCyan, "  %s URL:    %s", EmojiLink, r.URL)
		colorPrint(ColorCyan, "  %s Size:   %s", EmojiUpload, FormatSize(r.Size))
		if r.RemovalURL != "" {
			colorPrint(ColorYellow, "  %s Remove: %s", EmojiStop, r.RemovalURL)
		}
	}

	if len(allErrors) > 0 {
		fmt.Println()
		printError(msg.Errors)
		for _, e := range allErrors {
			colorPrint(ColorRed, "  %s %s", EmojiFail, e)
		}
	}

	if len(allResults) > 0 {
		firstURL := allResults[0].URL
		if err := CopyToClipboard(firstURL); err == nil {
			fmt.Println()
			printSuccess("%s %s", EmojiSuccess, msg.LinkCopied)
		}
	}

	fmt.Println()
	printInfo(msg.PressEnter)
	fmt.Scanln()
}

func getEnabledHosts(cfg Config) []string {
	var hosts []string
	for h, isOn := range cfg.HostsEnabled {
		if isOn {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

func uploadFile(hostName, filePath string, progress *Progress, cfg Config) (*host.UploadResult, error) {
	var h host.Host

	switch hostName {
	case "gofile.io":
		h = &host.Gofile{}
	case "1fichier.com":
		h = &host.Fichier{}
	case "fileditch.com":
		h = &host.FileDitch{}
	case "vikingfile.com":
		h = &host.VikingFile{}
	case "pixeldrain.com":
		h = &host.PixelDrain{APIKey: cfg.APIKeys["pixeldrain.com"]}
	case "buzzheavier.com":
		h = &host.BuzzHeavier{}
	case "litterbox.catbox.moe":
		h = &host.Litterbox{}
	case "filebin.net":
		h = &host.FileBin{}
	default:
		return nil, fmt.Errorf("unknown host: %s", hostName)
	}

 callback := func(current, total int64) {
		progress.Set(current)
	}

	result, err := h.Upload(filePath, host.UploadOptions{}, callback)
	if err != nil {
		return nil, err
	}

	progress.Done()
	return result, nil
}
