package splox

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSSEIterKeepalive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "data: keepalive")
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if !iter.Next() {
		t.Fatal("expected event")
	}
	if !iter.Event().IsKeepalive {
		t.Error("expected keepalive event")
	}
}

func TestSSEIterJSONEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"workflow_request":{"id":"req-1","workflow_version_id":"v1","start_node_id":"n1","status":"completed","created_at":"2025-01-01T00:00:00Z"}}`)
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if !iter.Next() {
		t.Fatal("expected event")
	}
	ev := iter.Event()
	if ev.WorkflowRequest == nil {
		t.Fatal("expected workflow_request")
	}
	if ev.WorkflowRequest.ID != "req-1" {
		t.Errorf("expected req-1, got %s", ev.WorkflowRequest.ID)
	}
	if ev.WorkflowRequest.Status != "completed" {
		t.Errorf("expected completed, got %s", ev.WorkflowRequest.Status)
	}
}

func TestSSEIterNodeExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"workflow_request":{"id":"req-1","workflow_version_id":"v1","start_node_id":"n1","status":"in_progress","created_at":"2025-01-01T00:00:00Z"},"node_execution":{"id":"ne-1","workflow_request_id":"req-1","node_id":"n1","workflow_version_id":"v1","status":"completed","output_data":{"text":"Hello world"}}}`)
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if !iter.Next() {
		t.Fatal("expected event")
	}
	ev := iter.Event()
	if ev.WorkflowRequest == nil {
		t.Fatal("expected workflow_request")
	}
	if ev.NodeExecution == nil {
		t.Fatal("expected node_execution")
	}
	if ev.NodeExecution.OutputData["text"] != "Hello world" {
		t.Errorf("expected Hello world, got %v", ev.NodeExecution.OutputData["text"])
	}
}

func TestSSEIterInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "data: {invalid json}")
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if !iter.Next() {
		t.Fatal("expected event")
	}
	ev := iter.Event()
	if ev.RawData != "{invalid json}" {
		t.Errorf("expected raw data, got %s", ev.RawData)
	}
	if ev.WorkflowRequest != nil {
		t.Error("expected nil workflow_request for invalid JSON")
	}
}

func TestSSEIterSkipsEmptyAndNonDataLines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, ": comment")
		fmt.Fprintln(w, "event: update")
		fmt.Fprintln(w, "data: keepalive")
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if !iter.Next() {
		t.Fatal("expected at least one event")
	}
	if !iter.Event().IsKeepalive {
		t.Error("first real event should be keepalive")
	}
}

func TestSSEIterMultipleEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "data: keepalive")
		fmt.Fprintln(w, `data: {"workflow_request":{"id":"req-1","workflow_version_id":"v1","start_node_id":"n1","status":"completed","created_at":"2025-01-01T00:00:00Z"}}`)
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	count := 0
	for iter.Next() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 events, got %d", count)
	}
	if err := iter.Err(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSSEIterStreamEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Empty — stream ends immediately
	}))
	defer srv.Close()

	client := NewClient("key", WithBaseURL(srv.URL))
	iter, err := client.streamSSE(t.Context(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if iter.Next() {
		t.Error("expected no events from empty stream")
	}
	if iter.Err() != nil {
		t.Errorf("unexpected error: %v", iter.Err())
	}
}
