package splox

import (
	"context"
	"fmt"
	"net/url"
)

// ChatService provides methods for the Chats API.
type ChatService struct {
	client *Client
}

// CreateChatParams are the parameters for [ChatService.Create].
type CreateChatParams struct {
	Name         string         `json:"name"`
	ResourceID   string         `json:"resource_id"`
	ResourceType string         `json:"resource_type,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Create creates a new chat session.
func (s *ChatService) Create(ctx context.Context, params CreateChatParams) (*Chat, error) {
	if params.ResourceType == "" {
		params.ResourceType = "api"
	}

	var resp Chat
	if err := s.client.do(ctx, "POST", "/chats", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get returns a chat session by ID.
func (s *ChatService) Get(ctx context.Context, chatID string) (*Chat, error) {
	var resp Chat
	if err := s.client.do(ctx, "GET", "/chats/"+chatID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListForResource returns all chats for a given resource.
func (s *ChatService) ListForResource(ctx context.Context, resourceType, resourceID string) (*ChatListResponse, error) {
	var resp ChatListResponse
	if err := s.client.do(ctx, "GET", "/chats/"+resourceType+"/"+resourceID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Listen opens an SSE stream for real-time chat events.
// The caller must call [SSEIter.Close] when done.
func (s *ChatService) Listen(ctx context.Context, chatID string) (*SSEIter, error) {
	return s.client.streamSSE(ctx, "/chat-internal-messages/"+chatID+"/listen")
}

// Delete removes a chat session.
func (s *ChatService) Delete(ctx context.Context, chatID string) error {
	return s.client.do(ctx, "DELETE", "/chats/"+chatID, nil, nil)
}

// ChatHistoryParams are optional parameters for [ChatService.GetHistory].
type ChatHistoryParams struct {
	Limit  int
	Before string // RFC3339 timestamp for backward pagination
}

// GetHistory returns paginated chat message history.
func (s *ChatService) GetHistory(ctx context.Context, chatID string, params *ChatHistoryParams) (*ChatHistoryResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Before != "" {
			v.Set("before", params.Before)
		}
	}

	var resp ChatHistoryResponse
	if err := s.client.do(ctx, "GET", addParams("/chat-history/"+chatID+"/paginated", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteHistory removes all message history for a chat.
func (s *ChatService) DeleteHistory(ctx context.Context, chatID string) error {
	return s.client.do(ctx, "DELETE", "/chat-history/"+chatID, nil, nil)
}
