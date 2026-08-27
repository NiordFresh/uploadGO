package host

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Litterbox struct{}

func (l *Litterbox) Name() string {
	return "litterbox.catbox.moe"
}

type litterboxResponse struct {
	URL string `json:"url"`
}

func (l *Litterbox) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("reqtype", "fileupload")
	writer.WriteField("time", "24h")

	part, err := writer.CreateFormFile("fileToUpload", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("cannot create form file: %w", err)
	}

	pw := &progressWriter{
		writer:   part,
		total:    info.Size(),
		callback: progress,
	}

	_, err = io.Copy(pw, file)
	if err != nil {
		return nil, fmt.Errorf("cannot write file to form: %w", err)
	}

	writer.Close()

	client := &http.Client{Timeout: 30 * time.Minute}
	req, err := http.NewRequest("POST", "https://litterbox.catbox.moe/resources/internals/api.php", body)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	url := string(respBody)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, url)
	}

	if url == "" || len(url) > 500 {
		return nil, fmt.Errorf("invalid response: %s", url)
	}

	return &UploadResult{
		URL:      url,
		Filename: filepath.Base(filePath),
		Size:     info.Size(),
	}, nil
}
