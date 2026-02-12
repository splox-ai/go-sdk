package splox_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	splox "github.com/splox-ai/go-sdk"
)

func mcpIntegrationClient(t *testing.T) *splox.Client {
	t.Helper()
	key := os.Getenv("SPLOX_API_KEY")
	if key == "" {
		t.Skip("SPLOX_API_KEY not set — skipping MCP integration test")
	}

	baseURL := os.Getenv("SPLOX_BASE_URL")
	if baseURL == "" {
		baseURL = splox.DefaultBaseURL
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	return splox.NewClient(key, splox.WithBaseURL(baseURL))
}

func TestMCPDiscoveryIntegration(t *testing.T) {
	client := mcpIntegrationClient(t)
	ctx := context.Background()

	searchQuery := os.Getenv("SPLOX_MCP_SEARCH_QUERY")

	catalog, err := client.MCP.ListCatalog(ctx, &splox.CatalogParams{Search: searchQuery, PerPage: 10})
	if err != nil {
		t.Fatalf("list catalog: %v", err)
	}
	if catalog.TotalCount < 0 {
		t.Fatalf("invalid total_count: %d", catalog.TotalCount)
	}

	servers, err := client.MCP.ListUserServers(ctx)
	if err != nil {
		t.Fatalf("list user servers: %v", err)
	}
	if servers.Total < 0 {
		t.Fatalf("invalid total: %d", servers.Total)
	}

	if len(servers.Servers) > 0 {
		tools, err := client.MCP.GetServerTools(ctx, servers.Servers[0].ID)
		if err != nil {
			t.Fatalf("get server tools: %v", err)
		}
		if tools.Total < 0 {
			t.Fatalf("invalid tools total: %d", tools.Total)
		}
	}
}

func TestMCPExecuteIntegration(t *testing.T) {
	client := mcpIntegrationClient(t)
	ctx := context.Background()

	serverID := os.Getenv("SPLOX_MCP_SERVER_ID")
	toolSlug := os.Getenv("SPLOX_MCP_TOOL_SLUG")
	if serverID == "" || toolSlug == "" {
		t.Skip("SPLOX_MCP_SERVER_ID and SPLOX_MCP_TOOL_SLUG not set — skipping MCP execute integration test")
	}

	argsJSON := os.Getenv("SPLOX_MCP_TOOL_ARGS_JSON")
	if argsJSON == "" {
		argsJSON = "{}"
	}

	args := map[string]any{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		t.Fatalf("invalid SPLOX_MCP_TOOL_ARGS_JSON: %v", err)
	}

	resp, err := client.MCP.ExecuteTool(ctx, splox.ExecuteToolParams{
		MCPServerID: serverID,
		ToolSlug:    toolSlug,
		Args:        args,
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}

	_ = resp.Result
}
