package splox

import "context"

// LLMService provides access to the Splox /chat/completions endpoint.
type LLMService struct {
	client *Client
}

// ChatParams are the parameters for a chat completion request.
type ChatParams struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
}

// Chat sends a chat completion request to the Splox LLM endpoint.
func (s *LLMService) Chat(ctx context.Context, params *ChatParams) (*ChatCompletion, error) {
	var resp ChatCompletion
	if err := s.client.do(ctx, "POST", "/chat/completions", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
