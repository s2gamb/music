package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// TelegramResponse represents the partial JSON structure we care about from Telegram.
// Removed explicit struct in favor of dynamic result map

// UploadAudio uploads an audio file to Telegram using sendAudio.
func UploadAudio(ctx context.Context, token, chatID, filePath, title, performer string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	writer.WriteField("chat_id", chatID)
	if title != "" {
		writer.WriteField("title", title)
	}
	if performer != "" {
		writer.WriteField("performer", performer)
	}
	writer.Close()

	return performTelegramUpload(ctx, token, "sendAudio", "audio", body, writer.FormDataContentType())
}

// UploadDocument uploads a non-audio file (like a cover) to Telegram using sendDocument.
func UploadDocument(ctx context.Context, token, chatID, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	writer.WriteField("chat_id", chatID)
	writer.Close()

	return performTelegramUpload(ctx, token, "sendDocument", "document", body, writer.FormDataContentType())
}

func performTelegramUpload(ctx context.Context, token, method, resultKey string, body *bytes.Buffer, contentType string) (string, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tr struct {
		Ok     bool                   `json:"ok"`
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &tr); err != nil {
		return "", err
	}

	if !tr.Ok {
		return "", fmt.Errorf("telegram api error: %s", string(respBody))
	}

	// Dynamic lookup based on the result key (audio or document)
	if res, ok := tr.Result[resultKey].(map[string]interface{}); ok {
		if fileID, ok := res["file_id"].(string); ok {
			return fileID, nil
		}
	}

	return "", fmt.Errorf("could not extract file_id from response")
}
