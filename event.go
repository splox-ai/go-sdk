package splox

import "context"

// EventService provides methods for the Events / Webhooks API.
type EventService struct {
	client *Client
}

// SendEventParams are the parameters for [EventService.Send].
type SendEventParams struct {
	WebhookID string
	Payload   map[string]any
	Secret    string // optional, sent as X-Webhook-Secret header
}

// Send triggers a workflow via webhook. No API key is required.
func (s *EventService) Send(ctx context.Context, params SendEventParams) (*EventResponse, error) {
	payload := params.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	if params.Secret != "" {
		fullURL := s.client.baseURL + "/events/" + params.WebhookID
		var resp EventResponse
		err := s.client.doWithHeaders(ctx, "POST", fullURL, payload, &resp, map[string]string{
			"X-Webhook-Secret": params.Secret,
		})
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	var resp EventResponse
	if err := s.client.do(ctx, "POST", "/events/"+params.WebhookID, payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
