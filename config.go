package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func GetConfigDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	dir := filepath.Join(appData, "uploadGO")
	os.MkdirAll(dir, 0755)
	return dir
}

func LoadSettings() Config {
	cfg := Config{
		Language: "pl",
		HostsEnabled: map[string]bool{
			"gofile.io":      true,
			"1fichier.com":  false,
			"fileditch.com":  true,
			"vikingfile.com": true,
			"pixeldrain.com": false,
			"buzzheavier.com": true,
		},
		APIKeys: map[string]string{},
	}

	exePath, _ := os.Executable()
	iniPath := filepath.Join(filepath.Dir(exePath), "settings.ini")

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return cfg
	}

	file, err := os.Open(iniPath)
	if err != nil {
		return cfg
	}
	defer file.Close()

	currentSection := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			section := strings.Trim(line, "[]")
			currentSection = strings.ToLower(section)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "settings":
			if key == "Language" {
				if value == "en" || value == "pl" {
					cfg.Language = value
				}
			}
		case "host":
			if value == "1" || strings.ToLower(value) == "true" {
				cfg.HostsEnabled[key] = true
			} else if value == "0" || strings.ToLower(value) == "false" {
				cfg.HostsEnabled[key] = false
			}
		case "apikeys":
			cfg.APIKeys[key] = value
		}
	}

	return cfg
}
