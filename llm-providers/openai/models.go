package openai

import (
	"net/http"
	"time"
)

type Client struct {
	apiKey      string
	orgID       string
	baseURL     string
	httpClient  *http.Client
	assistantID *string
}

// Assistant types
type Thread struct {
	ID string `json:"id"`
}

type Message struct {
	ID           string                 `json:"id"`
	Object       string                 `json:"object"`
	CreatedAt    time.Time              `json:"-"`          // Don't try to unmarshal directly
	RawCreatedAt int64                  `json:"created_at"` // Temporary field
	AssistantID  string                 `json:"assistant_id"`
	ThreadID     string                 `json:"thread_id"`
	RunID        string                 `json:"run_id"`
	Role         string                 `json:"role"`
	Content      []MessageContent       `json:"content"`
	Attachments  []interface{}          `json:"attachments"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type MessageContent struct {
	Type        string        `json:"type"`
	Text        MessageText   `json:"text"`
	Annotations []interface{} `json:"annotations"`
}

type MessageText struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations"`
}

type Run struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type MessageRequest struct {
	Role     string                 `json:"role"` // "user" or "assistant"
	Content  string                 `json:"content"`
	FileIDs  []string               `json:"file_ids,omitempty"` // For attachments
	Metadata map[string]interface{} `json:"metadata,omitempty"` // For custom data like timestamps
}

// OpenAIConfig holds config for OpenAI client
type OpenAIConfig struct {
	SecKey string
	OrgId  string
	AsstId *string // pointer: nil for classic, value for assistant
}
