package splox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer creates an httptest.Server that responds with the given status and body.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := NewClient("test-key", WithBaseURL(srv.URL))
	t.Cleanup(srv.Close)
	return srv, client
}

// --- Workflow tests ---

func TestWorkflowsList(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/workflows" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("search") != "my agent" {
			t.Errorf("expected search=my+agent, got %s", r.URL.Query().Get("search"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong auth header")
		}
		json.NewEncoder(w).Encode(WorkflowListResponse{
			Workflows: []Workflow{
				{ID: "wf-001", UserID: "user-001"},
			},
			Pagination: Pagination{Limit: 20, HasMore: false},
		})
	})

	resp, err := client.Workflows.List(context.Background(), &ListParams{
		Search: "my agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(resp.Workflows))
	}
	if resp.Workflows[0].ID != "wf-001" {
		t.Errorf("expected wf-001, got %s", resp.Workflows[0].ID)
	}
}

func TestWorkflowsGet(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workflows/wf-001" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(WorkflowFullResponse{
			Workflow:        Workflow{ID: "wf-001", UserID: "user-001"},
			WorkflowVersion: WorkflowVersion{ID: "ver-001", WorkflowID: "wf-001", Name: "Test", VersionNumber: 1, Status: "draft"},
			Nodes:           []Node{{ID: "n-001", WorkflowVersionID: "ver-001", NodeType: "start", Label: "Start"}},
			Edges:           []Edge{},
		})
	})

	resp, err := client.Workflows.Get(context.Background(), "wf-001")
	if err != nil {
		t.Fatal(err)
	}
	if resp.WorkflowVersion.Name != "Test" {
		t.Errorf("expected Test, got %s", resp.WorkflowVersion.Name)
	}
	if len(resp.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.Nodes))
	}
}

func TestWorkflowsGetLatestVersion(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workflows/wf-001/versions/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(WorkflowVersion{
			ID: "ver-002", WorkflowID: "wf-001", VersionNumber: 2, Name: "v2", Status: "published",
		})
	})

	v, err := client.Workflows.GetLatestVersion(context.Background(), "wf-001")
	if err != nil {
		t.Fatal(err)
	}
	if v.VersionNumber != 2 {
		t.Errorf("expected version 2, got %d", v.VersionNumber)
	}
}

func TestWorkflowsListVersions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(WorkflowVersionListResponse{
			Versions: []WorkflowVersion{
				{ID: "ver-001", WorkflowID: "wf-001", VersionNumber: 1, Name: "v1", Status: "draft"},
				{ID: "ver-002", WorkflowID: "wf-001", VersionNumber: 2, Name: "v2", Status: "published"},
			},
		})
	})

	resp, err := client.Workflows.ListVersions(context.Background(), "wf-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(resp.Versions))
	}
}

func TestWorkflowsGetStartNodes(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StartNodesResponse{
			Nodes: []Node{{ID: "n-001", WorkflowVersionID: "ver-001", NodeType: "start", Label: "Start"}},
		})
	})

	resp, err := client.Workflows.GetStartNodes(context.Background(), "ver-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 1 {
		t.Fatalf("expected 1 start node, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].Label != "Start" {
		t.Errorf("expected Start, got %s", resp.Nodes[0].Label)
	}
}

func TestWorkflowsRun(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body RunParams
		json.NewDecoder(r.Body).Decode(&body)
		if body.Query != "Hello" {
			t.Errorf("expected query Hello, got %s", body.Query)
		}
		if body.WorkflowVersionID != "ver-001" {
			t.Errorf("expected ver-001, got %s", body.WorkflowVersionID)
		}

		json.NewEncoder(w).Encode(RunResponse{WorkflowRequestID: "req-001"})
	})

	resp, err := client.Workflows.Run(context.Background(), RunParams{
		WorkflowVersionID: "ver-001",
		ChatID:            "chat-001",
		StartNodeID:       "node-001",
		Query:             "Hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.WorkflowRequestID != "req-001" {
		t.Errorf("expected req-001, got %s", resp.WorkflowRequestID)
	}
}

func TestWorkflowsRunWithFiles(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body RunParams
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(body.Files))
		}
		if body.Files[0].URL != "https://example.com/file.pdf" {
			t.Errorf("expected file URL, got %s", body.Files[0].URL)
		}
		json.NewEncoder(w).Encode(RunResponse{WorkflowRequestID: "req-002"})
	})

	resp, err := client.Workflows.Run(context.Background(), RunParams{
		WorkflowVersionID: "ver-001",
		ChatID:            "chat-001",
		StartNodeID:       "node-001",
		Query:             "Analyze this",
		Files: []WorkflowRequestFile{{
			URL:         "https://example.com/file.pdf",
			ContentType: "application/pdf",
			FileName:    "report.pdf",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.WorkflowRequestID != "req-002" {
		t.Errorf("expected req-002, got %s", resp.WorkflowRequestID)
	}
}

func TestWorkflowsGetExecutionTree(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ExecutionTreeResponse{
			ExecutionTree: ExecutionTree{
				WorkflowRequestID: "req-001",
				Status:            "completed",
				CreatedAt:         "2025-01-01T00:00:00Z",
				CompletedAt:       "2025-01-01T00:01:00Z",
				Nodes: []ExecutionNode{
					{
						ID:        "en-001",
						NodeID:    "node-001",
						Status:    "completed",
						NodeLabel: "Start",
						NodeType:  "start",
						OutputData: map[string]any{"text": "result"},
						ChildExecutions: []ChildExecution{
							{
								Index:             0,
								WorkflowRequestID: "req-002",
								Status:            "completed",
								Label:             "Agent B",
							},
						},
					},
				},
			},
		})
	})

	resp, err := client.Workflows.GetExecutionTree(context.Background(), "req-001")
	if err != nil {
		t.Fatal(err)
	}
	tree := resp.ExecutionTree
	if tree.Status != "completed" {
		t.Errorf("expected completed, got %s", tree.Status)
	}
	if len(tree.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].NodeLabel != "Start" {
		t.Errorf("expected Start, got %s", tree.Nodes[0].NodeLabel)
	}
	if len(tree.Nodes[0].ChildExecutions) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Nodes[0].ChildExecutions))
	}
	if tree.Nodes[0].ChildExecutions[0].Label != "Agent B" {
		t.Errorf("expected Agent B, got %s", tree.Nodes[0].ChildExecutions[0].Label)
	}
}

func TestWorkflowsGetHistory(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("expected limit=5, got %s", r.URL.Query().Get("limit"))
		}
		json.NewEncoder(w).Encode(HistoryResponse{
			Data: []WorkflowRequest{
				{ID: "req-001", WorkflowVersionID: "ver-001", StartNodeID: "n-001", Status: "completed", CreatedAt: "2025-01-01T00:00:00Z"},
			},
			Pagination: Pagination{Limit: 5, NextCursor: "req-000", HasMore: true},
		})
	})

	resp, err := client.Workflows.GetHistory(context.Background(), "req-001", &HistoryParams{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(resp.Data))
	}
	if resp.Pagination.NextCursor != "req-000" {
		t.Errorf("expected cursor req-000, got %s", resp.Pagination.NextCursor)
	}
	if !resp.Pagination.HasMore {
		t.Error("expected has_more=true")
	}
}

func TestWorkflowsStop(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/workflow-requests/req-001/stop" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Workflows.Stop(context.Background(), "req-001")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Chat tests ---

func TestChatsCreate(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/chats" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}

		var body CreateChatParams
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Test Chat" {
			t.Errorf("expected name Test Chat, got %s", body.Name)
		}
		if body.ResourceType != "api" {
			t.Errorf("expected resource_type api, got %s", body.ResourceType)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Chat{
			ID: "chat-001", Name: "Test Chat", ResourceType: "api", ResourceID: "wf-001",
		})
	})

	chat, err := client.Chats.Create(context.Background(), CreateChatParams{
		Name:       "Test Chat",
		ResourceID: "wf-001",
	})
	if err != nil {
		t.Fatal(err)
	}
	if chat.ID != "chat-001" {
		t.Errorf("expected chat-001, got %s", chat.ID)
	}
	if chat.Name != "Test Chat" {
		t.Errorf("expected Test Chat, got %s", chat.Name)
	}
}

func TestChatsGet(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats/chat-001" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Chat{
			ID: "chat-001", Name: "Test Chat",
		})
	})

	chat, err := client.Chats.Get(context.Background(), "chat-001")
	if err != nil {
		t.Fatal(err)
	}
	if chat.ID != "chat-001" {
		t.Errorf("expected chat-001, got %s", chat.ID)
	}
}

func TestChatsListForResource(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats/workflow/wf-001" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ChatListResponse{
			Chats: []Chat{
				{ID: "chat-001", Name: "Chat 1"},
				{ID: "chat-002", Name: "Chat 2"},
			},
		})
	})

	resp, err := client.Chats.ListForResource(context.Background(), "workflow", "wf-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Chats) != 2 {
		t.Fatalf("expected 2 chats, got %d", len(resp.Chats))
	}
}

func TestChatsGetHistory(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("before") != "2025-01-01T00:00:00Z" {
			t.Errorf("expected before param, got %s", r.URL.Query().Get("before"))
		}
		json.NewEncoder(w).Encode(ChatHistoryResponse{
			Messages: []ChatMessage{
				{
					ID: "msg-001", ChatID: "chat-001", Role: "user",
					Content: []ChatMessageContent{{Type: "text", Text: "Hello"}},
				},
			},
			HasMore: true,
		})
	})

	resp, err := client.Chats.GetHistory(context.Background(), "chat-001", &ChatHistoryParams{
		Limit:  10,
		Before: "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(resp.Messages))
	}
	if resp.Messages[0].Content[0].Text != "Hello" {
		t.Errorf("expected Hello, got %s", resp.Messages[0].Content[0].Text)
	}
	if !resp.HasMore {
		t.Error("expected has_more=true")
	}
}

func TestChatsDelete(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/chats/chat-001" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Chats.Delete(context.Background(), "chat-001")
	if err != nil {
		t.Fatal(err)
	}
}

func TestChatsDeleteHistory(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/chat-history/chat-001" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Chats.DeleteHistory(context.Background(), "chat-001")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Event tests ---

func TestEventsSend(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/events/wh-001" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "12345" {
			t.Errorf("expected order_id 12345, got %v", body["order_id"])
		}
		json.NewEncoder(w).Encode(EventResponse{OK: true, EventID: "evt-001"})
	})

	resp, err := client.Events.Send(context.Background(), SendEventParams{
		WebhookID: "wh-001",
		Payload:   map[string]any{"order_id": "12345"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.EventID != "evt-001" {
		t.Errorf("expected evt-001, got %s", resp.EventID)
	}
}

func TestEventsSendWithSecret(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Webhook-Secret") != "my-secret" {
			t.Errorf("expected X-Webhook-Secret: my-secret, got %s", r.Header.Get("X-Webhook-Secret"))
		}
		json.NewEncoder(w).Encode(EventResponse{OK: true, EventID: "evt-002"})
	})

	resp, err := client.Events.Send(context.Background(), SendEventParams{
		WebhookID: "wh-001",
		Payload:   map[string]any{"order": "456"},
		Secret:    "my-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.EventID != "evt-002" {
		t.Errorf("expected evt-002, got %s", resp.EventID)
	}
}

// --- Client config tests ---

func TestNewClientEnvFallback(t *testing.T) {
	t.Setenv("SPLOX_API_KEY", "env-key")
	client := NewClient("")
	if client.apiKey != "env-key" {
		t.Errorf("expected env-key, got %s", client.apiKey)
	}
}

func TestCustomBaseURL(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Chat{ID: "chat-001", Name: "Test"})
	})

	chat, err := client.Chats.Get(context.Background(), "chat-001")
	if err != nil {
		t.Fatal(err)
	}
	if chat.ID != "chat-001" {
		t.Errorf("expected chat-001, got %s", chat.ID)
	}
}
