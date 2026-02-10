package splox

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SSEIter reads Server-Sent Events from a stream.
// Call [SSEIter.Next] in a loop and [SSEIter.Close] when done.
type SSEIter struct {
	resp    *http.Response
	scanner *bufio.Scanner
	err     error
	event   SSEEvent
}

// Next advances to the next SSE event. Returns false when the stream
// ends or an error occurs (check [SSEIter.Err]).
func (it *SSEIter) Next() bool {
	for it.scanner.Scan() {
		line := strings.TrimSpace(it.scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		if payload == "keepalive" {
			it.event = SSEEvent{IsKeepalive: true, RawData: payload}
			return true
		}

		var ev SSEEvent
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			it.event = SSEEvent{RawData: payload}
			return true
		}

		ev.RawData = payload
		it.event = ev
		return true
	}

	if err := it.scanner.Err(); err != nil {
		it.err = &StreamError{Err: err}
	}
	return false
}

// Event returns the current SSE event. Only valid after [Next] returns true.
func (it *SSEIter) Event() SSEEvent {
	return it.event
}

// Err returns any error encountered during iteration.
func (it *SSEIter) Err() error {
	return it.err
}

// Close releases the underlying HTTP response.
func (it *SSEIter) Close() error {
	if it.resp != nil {
		return it.resp.Body.Close()
	}
	return nil
}

// streamSSE opens an SSE connection and returns an iterator.
func (c *Client) streamSSE(ctx context.Context, path string) (*SSEIter, error) {
	u := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("splox: create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Use a client without timeout for long-lived SSE streams.
	sseClient := &http.Client{Transport: c.httpClient.Transport}

	resp, err := sseClient.Do(req)
	if err != nil {
		return nil, &ConnectionError{Err: err}
	}

	if err := checkStatus(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return &SSEIter{
		resp:    resp,
		scanner: bufio.NewScanner(resp.Body),
	}, nil
}
