package host

type UploadOptions struct {
	Password string
	Mail     string
	Domain   string
}

type UploadResult struct {
	URL        string
	Filename   string
	Size       int64
	RemovalURL string
}

type ProgressCallback func(current, total int64)

type Host interface {
	Name() string
	Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error)
}
