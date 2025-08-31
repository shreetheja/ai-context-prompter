package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	llmproviders "github.com/shreetheja/ai-contextual-prompter/llm-providers"
)

// OpenAI LLM supports both Assistant API and classic completion/chat API.
// This package provides a unified interface for both modes.

const (
	openAiBase = "https://api.openai.com/v1"
)

// NewClient creates a new OpenAI client. If asstId is empty, classic mode is used.
func NewClient(secKey, orgId string, asstId *string) *Client {
	return &Client{
		apiKey:      secKey,
		orgID:       orgId,
		baseURL:     openAiBase,
		assistantID: asstId, // If empty, classic mode
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func New(cfg ...llmproviders.PromptOption) (*Client, error) {
	if len(cfg) == 0 {
		return nil, fmt.Errorf("no config provided for openai")
	}
	c, ok := cfg[0].(OpenAIConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for openai")
	}
	return NewClient(c.SecKey, c.OrgId, c.AsstId), nil
}

// =====================
// Assistant API Methods
// =====================
func (c *Client) CreateThread(ctx context.Context) (Thread, error) {
	resp, err := c.post(ctx, "threads", nil)
	if err != nil {
		return Thread{}, err
	}
	var thread Thread
	json.Unmarshal(resp, &thread)
	return thread, nil
}

func (c *Client) AddMessage(ctx context.Context, threadID string, msg MessageRequest) error {
	_, err := c.post(ctx, fmt.Sprintf("threads/%s/messages", threadID), msg)
	return err
}

func (c *Client) CreateRun(ctx context.Context, threadID, assistantID string) (Run, error) {
	resp, err := c.post(ctx, fmt.Sprintf("threads/%s/runs", threadID), map[string]string{
		"assistant_id": assistantID,
	})
	if err != nil {
		return Run{}, err
	}
	var run Run
	json.Unmarshal(resp, &run)
	return run, nil
}

func (c *Client) GetRun(ctx context.Context, threadID, runID string) (Run, error) {
	resp, err := c.get(ctx, fmt.Sprintf("threads/%s/runs/%s", threadID, runID))
	if err != nil {
		return Run{}, err
	}
	var run Run
	json.Unmarshal(resp, &run)
	return run, nil
}

func (c *Client) ListMessages(ctx context.Context, threadID string) ([]Message, error) {
	resp, err := c.get(ctx, fmt.Sprintf("threads/%s/messages", threadID))
	fmt.Printf("%v", string(resp))
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Message `json:"data"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	var data []Message
	for _, da := range result.Data {
		var cr = time.Unix(da.RawCreatedAt, 0)
		data = append(data, Message{
			Role:      da.Role,
			Content:   da.Content,
			CreatedAt: cr,
		})
	}
	return data, nil
}

// Embed returns the embedding vector for a given text using OpenAI's embedding API.
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	type embedReq struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	type embedResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	reqBody := embedReq{
		Model: "text-embedding-ada-002", // or configurable
		Input: []string{text},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Organization", c.orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	body, _ := io.ReadAll(resp.Body)
	var out embedResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return out.Data[0].Embedding, nil
}

// MaxContext returns the max context tokens for the model (hardcoded for now)
func (c *Client) MaxContext() int {
	return 16384 // gpt-4-32k, adjust as needed
}

// =====================
// Classic Completion/Chat API Methods
// =====================

// PromptClassic sends a prompt to the classic completion/chat API (no assistant).
func (c *Client) PromptClassic(ctx context.Context, prompt string, contextItems []string, opts ...llmproviders.PromptOption) (string, error) {
	// Example: Use gpt-3.5-turbo or gpt-4 chat API
	type chatReq struct {
		Model    string        `json:"model"`
		Messages []interface{} `json:"messages"`
	}
	type chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	// Build messages array: context as system/user, then prompt as user
	var messages []interface{}
	for _, ctxItem := range contextItems {
		messages = append(messages, map[string]string{"role": "system", "content": ctxItem})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	reqBody := chatReq{
		Model:    "gpt-3.5-turbo", // or configurable
		Messages: messages,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Organization", c.orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	body, _ := io.ReadAll(resp.Body)
	var out chatResp
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return out.Choices[0].Message.Content, nil
}

// PromptWithContext is a unified method to prompt either Assistant API or classic API based on config.
// If assistantID is set, uses Assistant API; otherwise, uses classic chat API.
func (c *Client) PromptWithContext(ctx context.Context, prompt string, contextItems []string, opts ...llmproviders.PromptOption) (string, error) {
	if c.assistantID != nil {
		// Use Assistant API: expects contextItems to be a sequence of messages (role/content)
		thread, err := c.CreateThread(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to create thread: %w", err)
		}
		threadID := thread.ID
		// Add all context messages to the thread
		for _, ctxItem := range contextItems {
			msg := MessageRequest{Role: "user", Content: ctxItem}
			if err := c.AddMessage(ctx, threadID, msg); err != nil {
				return "", fmt.Errorf("failed to add context message: %w", err)
			}
		}
		// Add user prompt
		if err := c.AddMessage(ctx, threadID, MessageRequest{Role: "user", Content: prompt}); err != nil {
			return "", fmt.Errorf("failed to add user prompt: %w", err)
		}
		assistant := ""
		if c.assistantID != nil {
			assistant = *c.assistantID
		}
		run, err := c.CreateRun(ctx, threadID, assistant)
		if err != nil {
			return "", fmt.Errorf("failed to create run: %w", err)
		}
		// Wait for completion (simple polling)
		for {
			runStatus, err := c.GetRun(ctx, threadID, run.ID)
			if err != nil {
				return "", fmt.Errorf("failed to get run status: %w", err)
			}
			if runStatus.Status == "completed" {
				break
			}
			if runStatus.Status == "failed" || runStatus.Status == "cancelled" || runStatus.Status == "expired" {
				return "", fmt.Errorf("run %s", runStatus.Status)
			}
			time.Sleep(1 * time.Second)
		}
		// Get the assistant's response
		messages, err := c.ListMessages(ctx, threadID)
		if err != nil {
			return "", fmt.Errorf("failed to get messages: %w", err)
		}
		if len(messages) == 0 {
			return "", fmt.Errorf("no messages returned")
		}
		lastMsg := messages[len(messages)-1]
		// If Message.Content is []MessageContent, join as string. If string, return directly.
		switch v := any(lastMsg.Content).(type) {
		case string:
			return v, nil
		case []MessageContent:
			var out string
			for _, mc := range v {
				out += mc.Text.Value
			}
			return out, nil
		default:
			return "", fmt.Errorf("unknown message content type")
		}
	} else {
		// Use classic chat API
		return c.PromptClassic(ctx, prompt, contextItems, opts...)
	}
}

// =====================
// Raw API's
// =====================

func (c *Client) delete(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("%s/%s", c.baseURL, endpoint),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Organization", c.orgID)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) post(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/%s", c.baseURL, endpoint),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}

	// Critical headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Organization", c.orgID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/%s", c.baseURL, endpoint),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Organization", c.orgID)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Name returns the name of the LLM provider.
func (c *Client) Name() string {
	return "openai"
}

var _ llmproviders.LLM = &Client{}
