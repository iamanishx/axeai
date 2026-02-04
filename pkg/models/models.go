package models

import (
	"time"
)

type ProviderType string

const (
	ProviderGemini ProviderType = "gemini"
	ProviderOpenAI ProviderType = "openai"
)

type Provider struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Type    ProviderType `json:"type"`
	APIKey  string       `json:"api_key"`
	Model   string       `json:"model"`
	Enabled bool         `json:"enabled"`
}

type MCPServerType string

const (
	MCPServerHTTP  MCPServerType = "http"
	MCPServerStdio MCPServerType = "stdio"
)

type MCPServer struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Type    MCPServerType `json:"type"`
	URL     string        `json:"url,omitempty"`
	Command string        `json:"command,omitempty"`
	Args    []string      `json:"args,omitempty"`
	Enabled bool          `json:"enabled"`
}

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

type MessageStatus string

const (
	StatusInProgress MessageStatus = "in_progress"
	StatusCompleted  MessageStatus = "completed"
	StatusCancelled  MessageStatus = "cancelled"
	StatusFailed     MessageStatus = "failed"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Session struct {
	ID           string     `db:"id" json:"id"`
	UserID       string     `db:"user_id" json:"user_id"`
	Title        string     `db:"title" json:"title"`
	Model        string     `db:"model" json:"model"`
	ProviderID   string     `db:"provider_id" json:"provider_id"`
	SystemPrompt string     `db:"system_prompt" json:"system_prompt"`
	Summary      *string    `db:"summary" json:"summary,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	ArchivedAt   *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type Message struct {
	ID         string         `db:"id" json:"id"`
	SessionID  string         `db:"session_id" json:"session_id"`
	Role       MessageRole    `db:"role" json:"role"`
	Content    string         `db:"content" json:"content"`
	Status     MessageStatus  `db:"status" json:"status"`
	TokenCount *int           `db:"token_count" json:"token_count,omitempty"`
	Metadata   map[string]any `db:"metadata_json" json:"metadata"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

type ToolCall struct {
	ID        string         `db:"id" json:"id"`
	SessionID string         `db:"session_id" json:"session_id"`
	MessageID string         `db:"message_id" json:"message_id"`
	ToolName  string         `db:"tool_name" json:"tool_name"`
	Args      map[string]any `db:"args_json" json:"args"`
	Result    map[string]any `db:"result_json" json:"result,omitempty"`
	Error     *string        `db:"error" json:"error,omitempty"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

type Attachment struct {
	ID        string         `db:"id" json:"id"`
	SessionID string         `db:"session_id" json:"session_id"`
	MessageID string         `db:"message_id" json:"message_id"`
	Type      string         `db:"type" json:"type"`
	Path      string         `db:"path" json:"path"`
	Metadata  map[string]any `db:"metadata_json" json:"metadata"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

type Settings struct {
	Key   string `db:"key" json:"key"`
	Value any    `db:"value_json" json:"value"`
}
