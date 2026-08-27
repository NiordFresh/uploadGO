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
		UploadingToAll:   "Wgrywanie na wszystkie włączone hostingi...",
		Results:          "Wyniki:",
		Errors:           "Błędy:",
		LinkCopied:       "Link skopiowany do schowka!",
		ErrorFileNotFound: "Plik nie został znaleziony:",
		ErrorNoHost:      "Żaden host nie jest włączony w settings.ini",
		Usage:            "Użycie: uploadGO.exe <plik> [--lang pl|en]",
		PressEnter:       "Naciśnij Enter aby kontynuować...",
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
