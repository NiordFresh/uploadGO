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

type Fichier struct{}

func (f *Fichier) Name() string {
	return "1fichier.com"
}

type fichierUploadServer struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

type fichierEndResponse struct {
	Incoming int `json:"incoming"`
	Links    []struct {
		Download string `json:"download"`
		Filename string `json:"filename"`
		Remove   string `json:"remove"`
		Size     string `json:"size"`
	} `json:"links"`
}

func (f *Fichier) Upload(filePath string, opts UploadOptions, progress ProgressCallback) (*UploadResult, error) {
	server, err := f.getUploadServer()
	if err != nil {
		return nil, fmt.Errorf("cannot get upload server: %w", err)
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

	part, err := writer.CreateFormFile("file[]", filepath.Base(filePath))
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

	if opts.Password != "" {
		writer.WriteField("dpass", opts.Password)
	}
	if opts.Mail != "" {
		writer.WriteField("mail", opts.Mail)
	}
	if opts.Domain != "" {
		writer.WriteField("domain", opts.Domain)
	}

	writer.Close()

	uploadURL := fmt.Sprintf("%s/upload.cgi?id=%s", server.URL, server.ID)

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	time.Sleep(2 * time.Second)

	endResp, err := f.getUploadResult(server.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot get upload result: %w", err)
	}

	if len(endResp.Links) == 0 {
		return nil, fmt.Errorf("no links in upload result")
	}

	link := endResp.Links[0]

	removalURL := ""
	if link.Remove != "" {
		removalURL = link.Remove
	}

	return &UploadResult{
		URL:        link.Download,
		Filename:   link.Filename,
		Size:       info.Size(),
		RemovalURL: removalURL,
	}, nil
}

func (f *Fichier) getUploadServer() (*fichierUploadServer, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.1fichier.com/api/upload.cgi")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var server fichierUploadServer
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, err
	}

	if server.URL == "" || server.ID == "" {
		return nil, fmt.Errorf("invalid server response")
	}

	return &server, nil
}

func (f *Fichier) getUploadResult(xid string) (*fichierEndResponse, error) {
	url := fmt.Sprintf("https://api.1fichier.com/api/end.pl?xid=%s&JSON=1", xid)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result fichierEndResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
