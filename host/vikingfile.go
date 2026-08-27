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

type VikingFile struct{}

func (v *VikingFile) Name() string {
	return "vikingfile.com"
}

type vikingServerResponse struct {
	Server string `json:"server"`
}

type vikingUploadResponse struct {
	Name string      `json:"name"`
	Size interface{} `json:"size"`
	Hash string      `json:"hash"`
	URL  string      `json:"url"`
}

func (v *VikingFile) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	server, err := v.getServer()
	if err != nil {
		return nil, fmt.Errorf("cannot get server: %w", err)
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

	writer.WriteField("user", "")
	writer.Close()

	client := &http.Client{Timeout: 30 * time.Minute}
	req, err := http.NewRequest("POST", server, body)
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

	var uploadResp vikingUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	if uploadResp.URL == "" {
		return nil, fmt.Errorf("upload failed: no URL in response")
	}

	var size int64
	switch v := uploadResp.Size.(type) {
	case float64:
		size = int64(v)
	case string:
		fmt.Sscanf(v, "%d", &size)
	}

	return &UploadResult{
		URL:      uploadResp.URL,
		Filename: uploadResp.Name,
		Size:     size,
	}, nil
}

func (v *VikingFile) getServer() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://vikingfile.com/api/get-server")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var serverResp vikingServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&serverResp); err != nil {
		return "", err
	}

	if serverResp.Server == "" {
		return "", fmt.Errorf("no server returned")
	}

	return serverResp.Server, nil
}
