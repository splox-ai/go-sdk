package splox_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	splox "github.com/splox-ai/go-sdk"
)

// Integration tests — run against the live Splox API.
//
// Requires environment variables:
//   SPLOX_API_KEY          — API token
//   SPLOX_BASE_URL         — (optional) defaults to https://app.splox.io/api/v1
//   SPLOX_TEST_WORKFLOW    — (optional) workflow name to search for, defaults to "Test"

func integrationClient(t *testing.T) *splox.Client {
	t.Helper()
	key := os.Getenv("SPLOX_API_KEY")
	if key == "" {
		t.Skip("SPLOX_API_KEY not set — skipping integration test")
	}
	baseURL := os.Getenv("SPLOX_BASE_URL")
	if baseURL == "" {
		baseURL = splox.DefaultBaseURL
	}
	return splox.NewClient(key, splox.WithBaseURL(baseURL))
}

func workflowName() string {
	name := os.Getenv("SPLOX_TEST_WORKFLOW")
	if name == "" {
		return "Test"
	}
	return name
}

func TestIntegration(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	wfName := workflowName()

	// 1. List workflows
	t.Log("1) Listing workflows...")
	wfList, err := client.Workflows.List(ctx, &splox.ListParams{Limit: 20})
	if err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	if len(wfList.Workflows) == 0 {
		t.Fatal("no workflows found")
	}
	t.Logf("   ✅ Found %d workflow(s)", len(wfList.Workflows))

	// 2. Search for workflow by name
	t.Logf("2) Searching for workflow %q...", wfName)
	searchResp, err := client.Workflows.List(ctx, &splox.ListParams{Search: wfName})
	if err != nil {
		t.Fatalf("search workflows: %v", err)
	}
	if len(searchResp.Workflows) == 0 {
		t.Fatalf("no workflow named %q found", wfName)
	}
	testWf := searchResp.Workflows[0]
	workflowID := testWf.ID
	t.Logf("   ✅ Found: %s", workflowID)

	// 3. Get workflow details
	t.Log("3) Getting workflow details...")
	wfFull, err := client.Workflows.Get(ctx, workflowID)
	if err != nil {
		t.Fatalf("get workflow: %v", err)
	}
	if wfFull.Workflow.ID != workflowID {
		t.Fatalf("expected %s, got %s", workflowID, wfFull.Workflow.ID)
	}
	t.Logf("   ✅ Got: %s (%d nodes, %d edges)", wfFull.WorkflowVersion.Name, len(wfFull.Nodes), len(wfFull.Edges))
	for _, n := range wfFull.Nodes {
		t.Logf("      • [%s] %s (%s)", n.NodeType, n.Label, n.ID[:12])
	}

	// 4. Get latest version
	t.Log("4) Getting latest version...")
	version, err := client.Workflows.GetLatestVersion(ctx, workflowID)
	if err != nil {
		t.Fatalf("get latest version: %v", err)
	}
	if version.WorkflowID != workflowID {
		t.Fatalf("expected workflow_id %s, got %s", workflowID, version.WorkflowID)
	}
	t.Logf("   ✅ v%d (status=%s, id=%s)", version.VersionNumber, version.Status, version.ID)

	// 5. List versions
	t.Log("5) Listing versions...")
	versionsResp, err := client.Workflows.ListVersions(ctx, workflowID)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versionsResp.Versions) == 0 {
		t.Fatal("no versions found")
	}
	t.Logf("   ✅ Found %d version(s)", len(versionsResp.Versions))

	// 6. Get start nodes
	t.Log("6) Getting start nodes...")
	startNodes, err := client.Workflows.GetStartNodes(ctx, version.ID)
	if err != nil {
		t.Fatalf("get start nodes: %v", err)
	}
	if len(startNodes.Nodes) == 0 {
		t.Fatal("no start nodes found")
	}
	startNode := startNodes.Nodes[0]
	t.Logf("   ✅ Found %d start node(s): %s — %s", len(startNodes.Nodes), startNode.ID, startNode.Label)

	// 7. Full programmatic flow: create chat → run → listen → get tree
	t.Log("7) Full programmatic flow (discover → run → wait)...")
	chat, err := client.Chats.Create(ctx, splox.CreateChatParams{
		Name:       "Go SDK Integration Test",
		ResourceID: workflowID,
	})
	if err != nil {
		t.Fatalf("create chat: %v", err)
	}
	t.Logf("   ▶ Chat created: %s", chat.ID)

	result, err := client.Workflows.Run(ctx, splox.RunParams{
		WorkflowVersionID: version.ID,
		ChatID:            chat.ID,
		StartNodeID:       startNode.ID,
		Query:             "Hello from Go SDK integration test!",
	})
	if err != nil {
		t.Fatalf("run workflow: %v", err)
	}
	t.Logf("   ▶ Started: %s", result.WorkflowRequestID)

	iter, err := client.Workflows.Listen(ctx, result.WorkflowRequestID)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	terminal := map[string]bool{"completed": true, "failed": true, "stopped": true}
	for iter.Next() {
		ev := iter.Event()
		if ev.IsKeepalive {
			continue
		}
		if ev.WorkflowRequest != nil && terminal[ev.WorkflowRequest.Status] {
			t.Logf("   ✅ Completed: %s", ev.WorkflowRequest.Status)
			break
		}
	}
	iter.Close()
	if err := iter.Err(); err != nil {
		t.Fatalf("listen error: %v", err)
	}

	treeResp, err := client.Workflows.GetExecutionTree(ctx, result.WorkflowRequestID)
	if err != nil {
		t.Fatalf("get execution tree: %v", err)
	}
	for _, node := range treeResp.ExecutionTree.Nodes {
		out := ""
		if node.OutputData != nil {
			text := fmt.Sprintf("%v", node.OutputData)
			if len(text) > 80 {
				text = text[:80] + "…"
			}
			out = " → " + text
		}
		t.Logf("      [%s] %s (%s)%s", node.Status, node.NodeLabel, node.NodeType, out)
	}

	// 8. Chat listen
	t.Log("8) Testing Chats.Listen() SSE...")
	result2, err := client.Workflows.Run(ctx, splox.RunParams{
		WorkflowVersionID: version.ID,
		ChatID:            chat.ID,
		StartNodeID:       startNode.ID,
		Query:             "Second message for chat listen",
	})
	if err != nil {
		t.Fatalf("run workflow 2: %v", err)
	}
	_ = result2

	chatIter, err := client.Chats.Listen(ctx, chat.ID)
	if err != nil {
		t.Fatalf("chat listen: %v", err)
	}
	chatEvents := 0
	for chatIter.Next() {
		ev := chatIter.Event()
		if ev.IsKeepalive {
			continue
		}
		chatEvents++
		if ev.WorkflowRequest != nil && terminal[ev.WorkflowRequest.Status] {
			break
		}
	}
	chatIter.Close()
	t.Logf("   ✅ Chat listen done — %d events", chatEvents)

	// 9. Stop workflow
	t.Log("9) Testing workflow stop...")
	result3, err := client.Workflows.Run(ctx, splox.RunParams{
		WorkflowVersionID: version.ID,
		ChatID:            chat.ID,
		StartNodeID:       startNode.ID,
		Query:             "This should be stopped",
	})
	if err != nil {
		t.Fatalf("run workflow 3: %v", err)
	}
	if err := client.Workflows.Stop(ctx, result3.WorkflowRequestID); err != nil {
		t.Logf("   ⚠️  Stop result: %v", err)
	} else {
		t.Logf("   ✅ Stop sent: %s", result3.WorkflowRequestID)
	}

	// 10. RunAndWait
	t.Log("10) Testing RunAndWait()...")
	chat2, err := client.Chats.Create(ctx, splox.CreateChatParams{
		Name:       "Go SDK RunAndWait",
		ResourceID: workflowID,
	})
	if err != nil {
		t.Fatalf("create chat2: %v", err)
	}
	treeResp2, err := client.Workflows.RunAndWait(ctx, splox.RunParams{
		WorkflowVersionID: version.ID,
		ChatID:            chat2.ID,
		StartNodeID:       startNode.ID,
		Query:             "Run and wait test from Go",
	}, 2*time.Minute)
	if err != nil {
		t.Fatalf("run_and_wait: %v", err)
	}
	if !terminal[treeResp2.ExecutionTree.Status] {
		t.Errorf("expected terminal status, got %s", treeResp2.ExecutionTree.Status)
	}
	t.Logf("   ✅ RunAndWait: %s", treeResp2.ExecutionTree.Status)

	// 11. Get execution history
	t.Log("11) Getting execution history...")
	history, err := client.Workflows.GetHistory(ctx, result.WorkflowRequestID, &splox.HistoryParams{Limit: 5})
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	t.Logf("   ✅ History: %d entries", len(history.Data))

	// 12. List chats for resource
	t.Log("12) Listing chats for workflow...")
	chatList, err := client.Chats.ListForResource(ctx, "api", workflowID)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	if len(chatList.Chats) == 0 {
		t.Fatal("expected at least 1 chat")
	}
	t.Logf("   ✅ Found %d chat(s)", len(chatList.Chats))

	// 13. Get chat message history
	t.Log("13) Getting chat message history...")
	chatHistory, err := client.Chats.GetHistory(ctx, chat.ID, &splox.ChatHistoryParams{Limit: 10})
	if err != nil {
		t.Fatalf("get chat history: %v", err)
	}
	t.Logf("   ✅ Got %d message(s) (has_more=%v)", len(chatHistory.Messages), chatHistory.HasMore)
	for i, msg := range chatHistory.Messages {
		if i >= 3 {
			break
		}
		preview := ""
		if len(msg.Content) > 0 && msg.Content[0].Text != "" {
			text := msg.Content[0].Text
			if len(text) > 60 {
				text = text[:60] + "…"
			}
			preview = " — " + text
		}
		t.Logf("      • [%s]%s", msg.Role, preview)
	}

	// 14. Delete chat history
	t.Log("14) Deleting chat history...")
	cleanupChat, err := client.Chats.Create(ctx, splox.CreateChatParams{
		Name:       "Go Cleanup Test",
		ResourceID: workflowID,
	})
	if err != nil {
		t.Fatalf("create cleanup chat: %v", err)
	}
	if err := client.Chats.DeleteHistory(ctx, cleanupChat.ID); err != nil {
		t.Fatalf("delete chat history: %v", err)
	}
	t.Logf("   ✅ Chat history deleted: %s", cleanupChat.ID)

	// 15. Delete chat
	t.Log("15) Deleting chat session...")
	if err := client.Chats.Delete(ctx, cleanupChat.ID); err != nil {
		t.Fatalf("delete chat: %v", err)
	}
	t.Logf("   ✅ Chat deleted: %s", cleanupChat.ID)

	t.Log("\n🎉 ALL INTEGRATION TESTS PASSED!")
}
