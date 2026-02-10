package splox

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckStatus401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer srv.Close()

	client := NewClient("bad-key", WithBaseURL(srv.URL))
	_, err := client.Chats.Get(t.Context(), "chat-001")
	if err == nil {
		t.Fatal("expected error")
	}

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T: %v", err, err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("expected 401, got %d", authErr.StatusCode)
	}
	if authErr.Message != "Invalid token" {
		t.Errorf("expected Invalid token, got %s", authErr.Message)
	}
}

func TestCheckStatus403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"Forbidden"}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	_, err := client.Chats.Get(t.Context(), "chat-001")

	var forbiddenErr *ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T", err)
	}
}

func TestCheckStatus404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Not found"}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	_, err := client.Chats.Get(t.Context(), "missing")

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestCheckStatus410(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(410)
		w.Write([]byte(`{"error":"Webhook expired"}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	_, err := client.Events.Send(t.Context(), SendEventParams{WebhookID: "wh-001"})

	var goneErr *GoneError
	if !errors.As(err, &goneErr) {
		t.Fatalf("expected GoneError, got %T", err)
	}
}

func TestCheckStatus429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(429)
		w.Write([]byte(`{"error":"Rate limit exceeded"}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	_, err := client.Workflows.List(t.Context(), nil)

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("expected RateLimitError, got %T", err)
	}
	if rlErr.RetryAfter != "60" {
		t.Errorf("expected RetryAfter=60, got %s", rlErr.RetryAfter)
	}
}

func TestCheckStatus500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Internal server error"}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	_, err := client.Chats.Get(t.Context(), "chat-001")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

func TestConnectionError(t *testing.T) {
	client := NewClient("key", WithBaseURL("http://localhost:1"))
	_, err := client.Chats.Get(t.Context(), "chat-001")

	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected ConnectionError, got %T: %v", err, err)
	}
}
