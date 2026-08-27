package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileDitch struct{}

func (f *FileDitch) Name() string {
	return "fileditch.com"
}

type fileditchResponse struct {
	Success  bool   `json:"success"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Error    string `json:"error"`
}

func (f *FileDitch) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
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

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
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
	req, err := http.NewRequest("POST", "https://new.fileditch.com/upload.php", body)
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

	var uploadResp fileditchResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	if !uploadResp.Success {
		errMsg := uploadResp.Error
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("upload failed: %s", errMsg)
	}

	return &UploadResult{
		URL:      uploadResp.URL,
		Filename: uploadResp.Filename,
		Size:     uploadResp.Size,
	}, nil
}
