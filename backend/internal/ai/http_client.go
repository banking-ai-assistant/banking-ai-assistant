package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type httpClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *httpClient) Analyze(ctx context.Context, content string) (*AnalyzeResponse, error) {
	body := map[string]string{
		"content": content,
	}
	raw, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analyze", bytes.NewBuffer(raw))
	if err != nil {
		return nil, fmt.Errorf("analyze: request build error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("analyze: http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("analyze: python returned status %d", resp.StatusCode)
	}

	var result AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("analyze: decode error: %w", err)
	}

	return &result, nil
}

func (c *httpClient) Generate(ctx context.Context, content, action string) (*GenerateResponse, error) {
	body := map[string]string{
		"content": content,
		"action":  action,
	}
	raw, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/generate", bytes.NewBuffer(raw))
	if err != nil {
		return nil, fmt.Errorf("generate: request build error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generate: http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("generate: python returned status %d", resp.StatusCode)
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("generate: decode error: %w", err)
	}

	return &result, nil
}
