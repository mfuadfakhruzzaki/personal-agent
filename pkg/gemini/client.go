package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	DefaultModel   = "gemini-1.5-flash"
)

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type GenerateRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerateResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type TodoItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = DefaultModel
	}

	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ExtractTodos(text string) ([]TodoItem, error) {
	prompt := c.buildPrompt(text)

	request := GenerateRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API returned status %d", resp.StatusCode)
	}

	var response GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	responseText := response.Candidates[0].Content.Parts[0].Text

	// Parse JSON response
	var todos []TodoItem
	if err := json.Unmarshal([]byte(responseText), &todos); err != nil {
		return nil, fmt.Errorf("failed to parse todos from response: %w", err)
	}

	return todos, nil
}

func (c *Client) buildPrompt(text string) string {
	return fmt.Sprintf(`Anda adalah asisten produktivitas. Dari teks berikut, ekstrak daftar todo dalam format JSON:
[{"title":"...","description":"...","due_date":"YYYY-MM-DD|null"}]

ATURAN:
1. Ekstrak hanya tugas/aktivitas yang perlu dilakukan
2. Jangan termasuk hal yang sudah selesai
3. due_date harus format YYYY-MM-DD atau null jika tidak ada tanggal
4. description boleh kosong jika tidak ada detail
5. Response harus valid JSON array

Teks:
---
%s`, text)
}
