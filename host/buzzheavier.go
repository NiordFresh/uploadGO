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

type BuzzHeavier struct{}

func (b *BuzzHeavier) Name() string {
	return "buzzheavier.com"
}

type buzzResponse struct {
	Code int `json:"code"`
	Data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"data"`
}

func (b *BuzzHeavier) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
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
	uploadURL := fmt.Sprintf("https://w.buzzheavier.com/%s", fileName)

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

	req, err := http.NewRequest("PUT", uploadURL, reader)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.ContentLength = info.Size()

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var buzzResp buzzResponse
	if err := json.Unmarshal(respBody, &buzzResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	if buzzResp.Data.ID == "" {
		return nil, fmt.Errorf("no file ID in response")
	}

	downloadURL := fmt.Sprintf("https://buzzheavier.com/f/%s", buzzResp.Data.ID)

	return &UploadResult{
		URL:      downloadURL,
		Filename: buzzResp.Data.Name,
		Size:     buzzResp.Data.Size,
	}, nil
}

type callbackReader struct {
	reader   io.Reader
	callback func(int)
}

func (cr *callbackReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	if n > 0 && cr.callback != nil {
		cr.callback(n)
	}
	return n, err
}
