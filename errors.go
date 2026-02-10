package splox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError is returned when the API responds with a non-2xx status code.
type APIError struct {
	StatusCode   int    `json:"-"`
	Message      string `json:"error"`
	ResponseBody string `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("splox: API error %d: %s", e.StatusCode, e.Message)
}

// AuthError is returned on 401 Unauthorized.
type AuthError struct{ APIError }

// ForbiddenError is returned on 403 Forbidden.
type ForbiddenError struct{ APIError }

// NotFoundError is returned on 404 Not Found.
type NotFoundError struct{ APIError }

// GoneError is returned on 410 Gone.
type GoneError struct{ APIError }

// RateLimitError is returned on 429 Too Many Requests.
type RateLimitError struct {
	APIError
	RetryAfter string // raw Retry-After header value
}

// ConnectionError is returned when the HTTP request fails at the transport level.
type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("splox: connection error: %v", e.Err)
}

func (e *ConnectionError) Unwrap() error { return e.Err }

// TimeoutError is returned when run-and-wait exceeds the deadline.
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("splox: timeout: %s", e.Message)
}

// StreamError is returned when SSE stream parsing fails.
type StreamError struct {
	Err error
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("splox: stream error: %v", e.Err)
}

func (e *StreamError) Unwrap() error { return e.Err }

// checkStatus inspects an HTTP response and returns a typed error for non-2xx.
func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	base := APIError{
		StatusCode:   resp.StatusCode,
		Message:      bodyStr,
		ResponseBody: bodyStr,
	}

	// Try to extract error message from JSON
	var parsed struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Error != "" {
		base.Message = parsed.Error
	}

	switch resp.StatusCode {
	case 401:
		return &AuthError{APIError: base}
	case 403:
		return &ForbiddenError{APIError: base}
	case 404:
		return &NotFoundError{APIError: base}
	case 410:
		return &GoneError{APIError: base}
	case 429:
		return &RateLimitError{
			APIError:   base,
			RetryAfter: resp.Header.Get("Retry-After"),
		}
	default:
		return &base
	}
}
