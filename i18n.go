package main

type Messages struct {
	UploadingToAll   string
	Results          string
	Errors           string
	LinkCopied       string
	ErrorFileNotFound string
	ErrorNoHost      string
	Usage            string
	PressEnter       string
}

var translations = map[string]Messages{
	"pl": {
		UploadingToAll:   "Wgrywanie na wszystkie wlaczone hostingi...",
		Results:          "Wyniki:",
		Errors:           "Bledy:",
		LinkCopied:       "Link skopiowany do schowka!",
		ErrorFileNotFound: "Plik nie zostal znaleziony:",
		ErrorNoHost:      "Zaden host nie jest wlaczony w settings.ini",
		Usage:            "Uzycie: uploadGO.exe <plik> [--lang pl|en]",
		PressEnter:       "Nacisnij Enter aby kontynuowac...",
	},
	"en": {
		UploadingToAll:   "Uploading to all enabled hosts...",
		Results:          "Results:",
		Errors:           "Errors:",
		LinkCopied:       "Link copied to clipboard!",
		ErrorFileNotFound: "File not found:",
		ErrorNoHost:      "No host enabled in settings.ini",
		Usage:            "Usage: uploadGO.exe <file> [--lang pl|en]",
		PressEnter:       "Press Enter to continue...",
	},
}

func GetMessages(lang string) Messages {
	if msg, ok := translations[lang]; ok {
		return msg
	}
	return translations["pl"]
}
