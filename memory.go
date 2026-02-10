package splox

import (
	"context"
	"fmt"
	"net/url"
)

// MemoryService provides methods for context memory operations.
type MemoryService struct {
	client *Client
}

// ── Models ───────────────────────────────────────────────────────────────────

// MemoryMessage represents a single message in node context memory.
type MemoryMessage struct {
	ID                string         `json:"id"`
	Role              string         `json:"role"`
	Content           any            `json:"content,omitempty"`
	ContextMemoryID   string         `json:"context_memory_id,omitempty"`
	AgentNodeID       string         `json:"agent_node_id,omitempty"`
	WorkflowVersionID string         `json:"workflow_version_id,omitempty"`
	ToolCalls         []map[string]any `json:"tool_calls,omitempty"`
	ToolCallID        string         `json:"tool_call_id,omitempty"`
	Files             []map[string]any `json:"files,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
}

// MemoryInstance represents a unique memory instance (context_memory_id + agent_node_id).
type MemoryInstance struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	WorkflowVersionID string `json:"workflow_version_id"`
	ChatID            string `json:"chat_id"`
	MemoryNodeID      string `json:"memory_node_id"`
	MemoryNodeLabel   string `json:"memory_node_label"`
	ContextSize       int    `json:"context_size"`
	MessageCount      int    `json:"message_count"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

// MemoryListResponse is returned by [MemoryService.List].
type MemoryListResponse struct {
	Chats      []MemoryInstance `json:"chats"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
}

// MemoryListParams are optional parameters for [MemoryService.List].
type MemoryListParams struct {
	Limit  int    // Instances per page (1-100, default 20)
	Cursor string // Pagination cursor from previous response
}

// MemoryGetResponse is returned by [MemoryService.Get].
type MemoryGetResponse struct {
	Messages   []MemoryMessage `json:"messages"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
	Limit      int             `json:"limit"`
}

// MemoryActionResponse is returned by summarize, trim, clear, and export actions.
type MemoryActionResponse struct {
	Action         string          `json:"action"`
	Message        string          `json:"message"`
	DeletedCount   int             `json:"deleted_count,omitempty"`
	Summary        string          `json:"summary,omitempty"`
	Messages       []MemoryMessage `json:"messages,omitempty"`
	RemainingCount int             `json:"remaining_count,omitempty"`
}

// ── Parameter types ──────────────────────────────────────────────────────────

// MemoryGetParams are parameters for [MemoryService.Get].
type MemoryGetParams struct {
	ChatID string // Required: the context memory ID (resolved chat/session ID)
	Limit  int    // Messages per page (1-100, default 20)
	Cursor string // Pagination cursor from previous response
}

// MemorySummarizeParams are parameters for [MemoryService.Summarize].
type MemorySummarizeParams struct {
	ContextMemoryID   string // Required
	WorkflowVersionID string // Required
	KeepLastN         *int   // Number of recent messages to keep
	SummarizePrompt   string // Custom prompt (falls back to agent config default)
}

// MemoryTrimParams are parameters for [MemoryService.Trim].
type MemoryTrimParams struct {
	ContextMemoryID   string // Required
	WorkflowVersionID string // Required
	MaxMessages       *int   // Maximum messages to keep (default 10)
}

// MemoryClearParams are parameters for [MemoryService.Clear].
type MemoryClearParams struct {
	ContextMemoryID   string // Required
	WorkflowVersionID string // Required
}

// MemoryExportParams are parameters for [MemoryService.Export].
type MemoryExportParams struct {
	ContextMemoryID   string // Required
	WorkflowVersionID string // Required
}

// MemoryDeleteParams are parameters for [MemoryService.Delete].
type MemoryDeleteParams struct {
	MemoryNodeID      string // Required: the agent/memory node ID
	WorkflowVersionID string // Required
}

// ── Methods ──────────────────────────────────────────────────────────────────

// List returns paginated memory instances for a workflow version.
func (s *MemoryService) List(ctx context.Context, workflowVersionID string, params *MemoryListParams) (*MemoryListResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Cursor != "" {
			v.Set("cursor", params.Cursor)
		}
	}

	var resp MemoryListResponse
	if err := s.client.do(ctx, "GET", addParams("/chat-memories/"+workflowVersionID, v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get returns paginated memory messages for an agent node.
func (s *MemoryService) Get(ctx context.Context, agentNodeID string, params *MemoryGetParams) (*MemoryGetResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.ChatID != "" {
			v.Set("chat_id", params.ChatID)
		}
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Cursor != "" {
			v.Set("cursor", params.Cursor)
		}
	}

	var resp MemoryGetResponse
	if err := s.client.do(ctx, "GET", addParams("/chat-memory/"+agentNodeID, v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Summarize replaces older memory messages with an LLM-generated summary.
func (s *MemoryService) Summarize(ctx context.Context, agentNodeID string, params MemorySummarizeParams) (*MemoryActionResponse, error) {
	body := map[string]any{
		"action":              "summarize",
		"context_memory_id":   params.ContextMemoryID,
		"workflow_version_id": params.WorkflowVersionID,
	}
	if params.KeepLastN != nil {
		body["keep_last_n"] = *params.KeepLastN
	}
	if params.SummarizePrompt != "" {
		body["summarize_prompt"] = params.SummarizePrompt
	}

	var resp MemoryActionResponse
	if err := s.client.do(ctx, "POST", "/chat-memory/"+agentNodeID+"/actions", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Trim removes the oldest messages to bring the count under maxMessages.
func (s *MemoryService) Trim(ctx context.Context, agentNodeID string, params MemoryTrimParams) (*MemoryActionResponse, error) {
	body := map[string]any{
		"action":              "trim",
		"context_memory_id":   params.ContextMemoryID,
		"workflow_version_id": params.WorkflowVersionID,
	}
	if params.MaxMessages != nil {
		body["max_messages"] = *params.MaxMessages
	}

	var resp MemoryActionResponse
	if err := s.client.do(ctx, "POST", "/chat-memory/"+agentNodeID+"/actions", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Clear removes all memory messages for a memory instance.
func (s *MemoryService) Clear(ctx context.Context, agentNodeID string, params MemoryClearParams) (*MemoryActionResponse, error) {
	body := map[string]any{
		"action":              "clear",
		"context_memory_id":   params.ContextMemoryID,
		"workflow_version_id": params.WorkflowVersionID,
	}

	var resp MemoryActionResponse
	if err := s.client.do(ctx, "POST", "/chat-memory/"+agentNodeID+"/actions", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Export returns all memory messages for a memory instance.
func (s *MemoryService) Export(ctx context.Context, agentNodeID string, params MemoryExportParams) (*MemoryActionResponse, error) {
	body := map[string]any{
		"action":              "export",
		"context_memory_id":   params.ContextMemoryID,
		"workflow_version_id": params.WorkflowVersionID,
	}

	var resp MemoryActionResponse
	if err := s.client.do(ctx, "POST", "/chat-memory/"+agentNodeID+"/actions", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes all memory for a specific memory instance.
func (s *MemoryService) Delete(ctx context.Context, contextMemoryID string, params MemoryDeleteParams) error {
	body := map[string]any{
		"memory_node_id":      params.MemoryNodeID,
		"workflow_version_id": params.WorkflowVersionID,
	}
	return s.client.do(ctx, "DELETE", "/chat-memories/"+contextMemoryID, body, nil)
}
