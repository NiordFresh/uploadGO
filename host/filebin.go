package host

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileBin struct{}

func (f *FileBin) Name() string {
	return "filebin.net"
}

type filebinResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Message  string `json:"message"`
}

func (f *FileBin) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	fileName := filepath.Base(filePath)
	binName := fmt.Sprintf("uploadgo-%d", time.Now().Unix())
	uploadURL := fmt.Sprintf("https://filebin.net/%s/%s", binName, fileName)

	var reader io.Reader = file
	if progress != nil {
		total := info.Size()
		current := int64(0)
		callback := progress
		reader = &callbackReader{
			reader: file,
			callback: func(n int) {
				current += int64(n)
				callback(current, total)
			},
		}
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	req, err := http.NewRequest("POST", uploadURL, reader)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.ContentLength = info.Size()
	req.Header.Set("filename", fileName)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var fileResp filebinResponse
	if err := json.Unmarshal(respBody, &fileResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	respFileName := fileResp.Filename
	if respFileName == "" {
		respFileName = fileName
	}

	downloadURL := fmt.Sprintf("https://filebin.net/%s/%s", binName, respFileName)

	return &UploadResult{
		URL:      downloadURL,
		Filename: respFileName,
		Size:     fileResp.Size,
	}, nil
}
