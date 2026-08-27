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

type Gofile struct{}

func (g *Gofile) Name() string {
	return "gofile.io"
}

type gofileServersResponse struct {
	Status string `json:"status"`
	Data   struct {
		Servers []struct {
			Name string `json:"name"`
			Zone string `json:"zone"`
		} `json:"servers"`
	} `json:"data"`
}

type gofileUploadResponse struct {
	Status string `json:"status"`
	Data   struct {
		Code         string `json:"code"`
		DownloadPage string `json:"downloadPage"`
		FileName     string `json:"fileName"`
		FileID       string `json:"fileId"`
		RemovalCode  string `json:"removalCode"`
	} `json:"data"`
}

type progressWriter struct {
	writer   io.Writer
	current  int64
	total    int64
	callback func(current, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.current += int64(n)
		if pw.callback != nil {
			pw.callback(pw.current, pw.total)
		}
	}
	return n, err
}

func (g *Gofile) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	fileSize := info.Size()

	server, err := g.getBestServer()
	if err != nil {
		return nil, fmt.Errorf("cannot get server: %w", err)
	}

	uploadURL := fmt.Sprintf("https://%s.gofile.io/uploadFile", server)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("cannot create form file: %w", err)
	}

	pw := &progressWriter{
		writer:   part,
		total:    fileSize,
		callback: progress,
	}

	_, err = io.Copy(pw, file)
	if err != nil {
		return nil, fmt.Errorf("cannot write file to form: %w", err)
	}

	writer.Close()

	client := &http.Client{Timeout: 30 * time.Minute}
	req, err := http.NewRequest("POST", uploadURL, body)
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

	var uploadResp gofileUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %s", string(respBody))
	}

	if uploadResp.Status != "ok" {
		return nil, fmt.Errorf("upload failed: %s", uploadResp.Status)
	}

	removalURL := ""
	if uploadResp.Data.RemovalCode != "" {
		removalURL = fmt.Sprintf("https://gofile.io/?c=%s&rm=%s",
			uploadResp.Data.Code, uploadResp.Data.RemovalCode)
	}

	return &UploadResult{
		URL:        uploadResp.Data.DownloadPage,
		Filename:   uploadResp.Data.FileName,
		Size:       fileSize,
		RemovalURL: removalURL,
	}, nil
}

func (g *Gofile) getBestServer() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.gofile.io/servers")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var serverResp gofileServersResponse
	if err := json.NewDecoder(resp.Body).Decode(&serverResp); err != nil {
		return "", err
	}

	if serverResp.Status != "ok" || len(serverResp.Data.Servers) == 0 {
		return "", fmt.Errorf("no servers available")
	}

	return serverResp.Data.Servers[0].Name, nil
}
