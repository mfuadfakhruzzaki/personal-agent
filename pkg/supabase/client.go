package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"todo-agent-backend/internal/models"
)

type Client struct {
	url        string
	key        string
	httpClient *http.Client
}

func NewClient(url, key string) *Client {
	return &Client{
		url: url,
		key: key,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) InsertTodo(todo *models.Todo) error {
	url := fmt.Sprintf("%s/rest/v1/todos", c.url)

	jsonData, err := json.Marshal(todo)
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.key)
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) InsertTodos(todos []models.Todo) error {
	if len(todos) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/rest/v1/todos", c.url)

	jsonData, err := json.Marshal(todos)
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.key)
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetTodosByUserID(userID string) ([]models.Todo, error) {
	url := fmt.Sprintf("%s/rest/v1/todos?user_id=eq.%s&order=created_at.desc", c.url, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.key)
	req.Header.Set("Authorization", "Bearer "+c.key)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}

	var todos []models.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return nil, fmt.Errorf("failed to decode todos: %w", err)
	}

	return todos, nil
}
