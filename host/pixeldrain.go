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

type PixelDrain struct {
	APIKey string
}

func (p *PixelDrain) Name() string {
	return "pixeldrain.com"
}

type pixeldrainResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (p *PixelDrain) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("API key required (set in settings.ini)")
	}

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
	req, err := http.NewRequest("POST", "https://pixeldrain.com/api/file", body)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("", p.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var uploadResp pixeldrainResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	if !uploadResp.Success {
		errMsg := uploadResp.Message
		if errMsg == "" {
			errMsg = uploadResp.Value
		}
		return nil, fmt.Errorf("upload failed: %s", errMsg)
	}

	downloadURL := fmt.Sprintf("https://pixeldrain.com/u/%s", uploadResp.ID)

	return &UploadResult{
		URL:      downloadURL,
		Filename: filepath.Base(filePath),
		Size:     info.Size(),
	}, nil
}
