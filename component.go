package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Language     string
	HostsEnabled map[string]bool
	APIKeys      map[string]string
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CopyToClipboard(text string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "powershell"
		args = []string{"-Command", fmt.Sprintf("Set-Clipboard -Value '%s'", text)}
	case "darwin":
		cmd = "pbcopy"
		args = []string{text}
	default:
		cmd = "xclip"
		args = []string{"-selection", "clipboard"}
	}

	return ExecCommand(cmd, args...)
}

func ExecCommand(name string, args ...string) error {
	cmd := createCommand(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func GetExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func GetRegFilePath() string {
	return filepath.Join(GetExeDir(), "uploadGO.reg")
}

func GetRemoveRegFilePath() string {
	return filepath.Join(GetExeDir(), "uploadGO_remove.reg")
}

func GenerateRegFiles() error {
	exePath := filepath.Join(GetExeDir(), "uploadGO.exe")
	exePath = strings.ReplaceAll(exePath, `\`, `\\`)

	regContent := fmt.Sprintf(`Windows Registry Editor Version 5.00

[HKEY_CLASSES_ROOT\*\shell\uploadGO]
@="Upload with uploadGO"
"Icon"="%s"

[HKEY_CLASSES_ROOT\*\shell\uploadGO\command]
@="\"%s\" \"%%1\""
`, exePath, exePath)

 removeContent := `Windows Registry Editor Version 5.00

[-HKEY_CLASSES_ROOT\*\shell\uploadGO]
`

	if err := os.WriteFile(GetRegFilePath(), []byte(regContent), 0644); err != nil {
		return err
	}

	return os.WriteFile(GetRemoveRegFilePath(), []byte(removeContent), 0644)
}
