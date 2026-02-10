# Splox Go SDK

Official Go client for the [Splox API](https://docs.splox.io) — run workflows, manage chats, browse the MCP catalog, and monitor execution programmatically.

[![Go Reference](https://pkg.go.dev/badge/github.com/splox-ai/go-sdk.svg)](https://pkg.go.dev/github.com/splox-ai/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Installation

```bash
go get github.com/splox-ai/go-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	splox "github.com/splox-ai/go-sdk"
)

func main() {
	ctx := context.Background()
	client := splox.NewClient("your-api-key")

	// Discover workflows
	workflows, err := client.Workflows.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	wf := workflows.Workflows[0]
	fmt.Println("Workflow:", wf.LatestVersion.Name)

	// Get full details
	full, _ := client.Workflows.Get(ctx, wf.ID)
	version := full.WorkflowVersion

	// Find start node
	startNodes, _ := client.Workflows.GetStartNodes(ctx, version.ID)
	startNode := startNodes.Nodes[0]

	// Create a chat session
	chat, _ := client.Chats.Create(ctx, splox.CreateChatParams{
		Name:       "My Session",
		ResourceID: wf.ID,
	})

	// Run the workflow
	result, _ := client.Workflows.Run(ctx, splox.RunParams{
		WorkflowVersionID: version.ID,
		ChatID:            chat.ID,
		StartNodeID:       startNode.ID,
		Query:             "Summarize the latest sales report",
	})

	// Get execution tree
	tree, _ := client.Workflows.GetExecutionTree(ctx, result.WorkflowRequestID)
	for _, node := range tree.ExecutionTree.Nodes {
		fmt.Printf("%s (%s): %v\n", node.NodeLabel, node.Status, node.OutputData)
	}
}
```

## Client Options

```go
// API key from environment variable (SPLOX_API_KEY)
client := splox.NewClient("")

// Custom base URL (self-hosted)
client := splox.NewClient("key", splox.WithBaseURL("https://your-instance.com/api/v1"))

// Custom HTTP client
client := splox.NewClient("key", splox.WithHTTPClient(&http.Client{
	Transport: &http.Transport{TLSClientConfig: tlsConfig},
}))

// Custom timeout
client := splox.NewClient("key", splox.WithTimeout(60*time.Second))
```

## Streaming (SSE)

### Listen to workflow execution

```go
iter, err := client.Workflows.Listen(ctx, workflowRequestID)
if err != nil {
	log.Fatal(err)
}
defer iter.Close()

for iter.Next() {
	ev := iter.Event()
	if ev.IsKeepalive {
		continue
	}
	if ev.NodeExecution != nil {
		fmt.Printf("[%s] %s\n", ev.NodeExecution.Status, ev.NodeExecution.NodeID)
	}
	if ev.WorkflowRequest != nil && ev.WorkflowRequest.Status == "completed" {
		break
	}
}
if err := iter.Err(); err != nil {
	log.Fatal(err)
}
```

### Listen to chat messages

Stream real-time chat events including text deltas, tool calls, and more:

```go
iter, err := client.Chats.Listen(ctx, chatID)
if err != nil {
	log.Fatal(err)
}
defer iter.Close()

var response strings.Builder
for iter.Next() {
	ev := iter.Event()
	if ev.IsKeepalive {
		continue
	}

	switch ev.EventType {
	case "text_delta":
		fmt.Print(ev.TextDelta)
		response.WriteString(ev.TextDelta)
	case "tool_call_start":
		fmt.Printf("\nCalling tool: %s\n", ev.ToolName)
	case "done":
		fmt.Println("\nIteration complete")
	}

	// Stop when workflow completes
	if ev.WorkflowRequest != nil && ev.WorkflowRequest.Status == "completed" {
		break
	}
}

fmt.Println("Final response:", response.String())
```

**Event types:**

| Type | Fields | Description |
|------|--------|-------------|
| `text_delta` | `TextDelta` | Streamed text chunk |
| `reasoning_delta` | `ReasoningDelta`, `ReasoningType` | Thinking content |
| `tool_call_start` | `ToolCallID`, `ToolName` | Tool call initiated |
| `tool_call_delta` | `ToolCallID`, `ToolArgsDelta` | Tool arguments delta |
| `tool_complete` | `ToolName`, `ToolCallID`, `ToolResult` | Tool finished |
| `tool_error` | `ToolName`, `ToolCallID`, `Error` | Tool failed |
| `done` | `Iteration`, `RunID` | Iteration complete |
| `error` | `Error` | Error occurred |

## Run & Wait

Blocks until the workflow reaches a terminal state:

```go
tree, err := client.Workflows.RunAndWait(ctx, splox.RunParams{
	WorkflowVersionID: versionID,
	ChatID:            chatID,
	StartNodeID:       startNodeID,
	Query:             "Process this request",
}, 5*time.Minute)
if err != nil {
	log.Fatal(err)
}

fmt.Println("Status:", tree.ExecutionTree.Status)
for _, node := range tree.ExecutionTree.Nodes {
	fmt.Printf("  [%s] %s: %v\n", node.Status, node.NodeLabel, node.OutputData)
}
```

## Workflows

```go
// List with search and pagination
resp, _ := client.Workflows.List(ctx, &splox.ListParams{
	Limit:  10,
	Search: "my agent",
})

// Get workflow details (nodes, edges, version)
full, _ := client.Workflows.Get(ctx, "workflow-id")

// List all versions
versions, _ := client.Workflows.ListVersions(ctx, "workflow-id")

// Get latest version
version, _ := client.Workflows.GetLatestVersion(ctx, "workflow-id")

// Get start nodes
startNodes, _ := client.Workflows.GetStartNodes(ctx, "workflow-version-id")

// Run with file attachments
result, _ := client.Workflows.Run(ctx, splox.RunParams{
	WorkflowVersionID: "version-id",
	ChatID:            "chat-id",
	StartNodeID:       "node-id",
	Query:             "Hello!",
	Files: []splox.WorkflowRequestFile{{
		URL:         "https://example.com/doc.pdf",
		ContentType: "application/pdf",
		FileName:    "report.pdf",
	}},
})

// Stop execution
_ = client.Workflows.Stop(ctx, "workflow-request-id")

// Get execution tree
tree, _ := client.Workflows.GetExecutionTree(ctx, "workflow-request-id")

// Get execution history
history, _ := client.Workflows.GetHistory(ctx, "workflow-request-id", &splox.HistoryParams{
	Limit: 25,
})
```

## Chats

```go
// Create
chat, _ := client.Chats.Create(ctx, splox.CreateChatParams{
	Name:         "Support Session",
	ResourceID:   "workflow-id",
	ResourceType: "api",
	Metadata:     map[string]any{"user": "123"},
})

// Get
chat, _ := client.Chats.Get(ctx, "chat-id")

// List for a resource
list, _ := client.Chats.ListForResource(ctx, "workflow", "workflow-id")

// Get message history with pagination
history, _ := client.Chats.GetHistory(ctx, "chat-id", &splox.ChatHistoryParams{
	Limit: 50,
})
if history.HasMore {
	oldest := history.Messages[len(history.Messages)-1]
	older, _ := client.Chats.GetHistory(ctx, "chat-id", &splox.ChatHistoryParams{
		Before: oldest.CreatedAt,
	})
}

// Delete history
_ = client.Chats.DeleteHistory(ctx, "chat-id")

// Delete chat
_ = client.Chats.Delete(ctx, "chat-id")
```

## Memory

Inspect and manage agent context memory — list instances, read messages, summarize, trim, clear, or export.

```go
// List memory instances for a workflow version (paginated)
instances, _ := client.Memory.List(ctx, "workflow-version-id", &splox.MemoryListParams{
	Limit: 20,
})
for _, inst := range instances.Chats {
	fmt.Printf("%s: %d messages\n", inst.MemoryNodeLabel, inst.MessageCount)
}
// Paginate
if instances.HasMore {
	next, _ := client.Memory.List(ctx, "workflow-version-id", &splox.MemoryListParams{
		Cursor: instances.NextCursor,
	})
}

// Get memory messages for an agent node
messages, _ := client.Memory.Get(ctx, "agent-node-id", &splox.MemoryGetParams{
	ChatID: "session-id",
	Limit:  20,
})
for _, msg := range messages.Messages {
	fmt.Printf("[%s] %v\n", msg.Role, msg.Content)
}

// Summarize — compress older messages into an LLM-generated summary
keepN := 3
sumResult, _ := client.Memory.Summarize(ctx, "agent-node-id", splox.MemorySummarizeParams{
	ContextMemoryID:   "session-id",
	WorkflowVersionID: "version-id",
	KeepLastN:         &keepN,
})
fmt.Println("Summary:", sumResult.Summary)

// Trim — drop oldest messages to stay under a limit
maxMsgs := 20
client.Memory.Trim(ctx, "agent-node-id", splox.MemoryTrimParams{
	ContextMemoryID:   "session-id",
	WorkflowVersionID: "version-id",
	MaxMessages:       &maxMsgs,
})

// Export — get all messages without modifying them
exported, _ := client.Memory.Export(ctx, "agent-node-id", splox.MemoryExportParams{
	ContextMemoryID:   "session-id",
	WorkflowVersionID: "version-id",
})

// Clear — remove all messages
client.Memory.Clear(ctx, "agent-node-id", splox.MemoryClearParams{
	ContextMemoryID:   "session-id",
	WorkflowVersionID: "version-id",
})

// Delete a specific memory instance
client.Memory.Delete(ctx, "session-id", &splox.MemoryDeleteParams{
	MemoryNodeID:      "agent-node-id",
	WorkflowVersionID: "version-id",
})
```

## MCP (Model Context Protocol)

Browse the MCP server catalog, manage end-user connections, and generate credential-submission links.

### Catalog

```go
// Search the MCP catalog
catalog, _ := client.MCP.ListCatalog(ctx, &splox.CatalogParams{
	Search:  "github",
	PerPage: 10,
})
for _, server := range catalog.MCPServers {
	fmt.Printf("%s — %s\n", server.Name, server.URL)
}

// Get featured servers
featured, _ := client.MCP.ListCatalog(ctx, &splox.CatalogParams{Featured: true})

// Get a single catalog item
item, _ := client.MCP.GetCatalogItem(ctx, "mcp-server-id")
fmt.Println(item.Name, item.AuthType)
```

### Connections

```go
// List all end-user connections
conns, _ := client.MCP.ListConnections(ctx, nil)
fmt.Printf("%d connections\n", conns.Total)

// Filter by MCP server or end-user
conns, _ = client.MCP.ListConnections(ctx, &splox.ConnectionParams{
	MCPServerID: "server-id",
	EndUserID:   "user-123",
})

// Delete a connection
_ = client.MCP.DeleteConnection(ctx, "connection-id")
```

### Connection Token & Link

Generate signed JWTs for end-user credential submission — no API call required:

```go
// Generate a token (expires in 1 hour)
token, _ := splox.GenerateConnectionToken(
	"mcp-server-id",
	"owner-user-id",
	"end-user-id",
	"your-credentials-encryption-key",
)

// Generate a full connection link
link, _ := splox.GenerateConnectionLink(
	"https://app.splox.io",
	"mcp-server-id",
	"owner-user-id",
	"end-user-id",
	"your-credentials-encryption-key",
)
// → https://app.splox.io/tools/connect?token=eyJhbG...
```

## Webhooks

```go
// No API key required
client := splox.NewClient("")

resp, _ := client.Events.Send(ctx, splox.SendEventParams{
	WebhookID: "your-webhook-id",
	Payload:   map[string]any{"order_id": "12345", "status": "paid"},
	Secret:    "optional-webhook-secret",
})
fmt.Println(resp.EventID)
```

## Error Handling

```go
import "errors"

result, err := client.Workflows.Run(ctx, params)
if err != nil {
	var authErr *splox.AuthError
	var notFound *splox.NotFoundError
	var rateLimit *splox.RateLimitError
	var apiErr *splox.APIError
	var timeoutErr *splox.TimeoutError

	switch {
	case errors.As(err, &authErr):
		log.Fatal("Invalid API key")
	case errors.As(err, &notFound):
		log.Fatal("Resource not found")
	case errors.As(err, &rateLimit):
		log.Fatalf("Rate limited, retry after %s", rateLimit.RetryAfter)
	case errors.As(err, &timeoutErr):
		log.Fatal("Operation timed out")
	case errors.As(err, &apiErr):
		log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
	default:
		log.Fatal(err)
	}
}
```

## Full API Reference

### `client.Workflows`

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, *ListParams)` | `*WorkflowListResponse` | List workflows with pagination |
| `Get(ctx, workflowID)` | `*WorkflowFullResponse` | Get workflow with nodes, edges, version |
| `GetLatestVersion(ctx, workflowID)` | `*WorkflowVersion` | Get latest version |
| `ListVersions(ctx, workflowID)` | `*WorkflowVersionListResponse` | List all versions |
| `GetStartNodes(ctx, versionID)` | `*StartNodesResponse` | Get start nodes |
| `Run(ctx, RunParams)` | `*RunResponse` | Trigger execution |
| `Listen(ctx, requestID)` | `*SSEIter` | Stream execution events |
| `GetExecutionTree(ctx, requestID)` | `*ExecutionTreeResponse` | Get execution hierarchy |
| `GetHistory(ctx, requestID, *HistoryParams)` | `*HistoryResponse` | Paginated execution history |
| `Stop(ctx, requestID)` | `error` | Stop a running execution |
| `RunAndWait(ctx, RunParams, timeout)` | `*ExecutionTreeResponse` | Run and wait for completion |

### `client.Chats`

| Method | Returns | Description |
|--------|---------|-------------|
| `Create(ctx, CreateChatParams)` | `*Chat` | Create a chat session |
| `Get(ctx, chatID)` | `*Chat` | Get chat by ID |
| `ListForResource(ctx, type, id)` | `*ChatListResponse` | List chats for a resource |
| `Listen(ctx, chatID)` | `*SSEIter` | Stream chat events |
| `GetHistory(ctx, chatID, *ChatHistoryParams)` | `*ChatHistoryResponse` | Paginated message history |
| `DeleteHistory(ctx, chatID)` | `error` | Delete all messages |
| `Delete(ctx, chatID)` | `error` | Delete chat session |

### `client.Events`

| Method | Returns | Description |
|--------|---------|-------------|
| `Send(ctx, SendEventParams)` | `*EventResponse` | Send event via webhook |

### `client.Memory`

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, versionID, *MemoryListParams)` | `*MemoryListResponse` | List memory instances (paginated) |
| `Get(ctx, nodeID, *MemoryGetParams)` | `*MemoryGetResponse` | Get paginated messages |
| `Summarize(ctx, nodeID, MemorySummarizeParams)` | `*MemoryActionResponse` | Summarize older messages |
| `Trim(ctx, nodeID, MemoryTrimParams)` | `*MemoryActionResponse` | Drop oldest messages |
| `Clear(ctx, nodeID, MemoryClearParams)` | `*MemoryActionResponse` | Remove all messages |
| `Export(ctx, nodeID, MemoryExportParams)` | `*MemoryActionResponse` | Export all messages |
| `Delete(ctx, memoryID, *MemoryDeleteParams)` | `error` | Delete a memory instance |

### `client.MCP`

| Method | Returns | Description |
|--------|---------|-------------|
| `ListCatalog(ctx, *CatalogParams)` | `*MCPCatalogListResponse` | Search/list MCP catalog (paginated) |
| `GetCatalogItem(ctx, id)` | `*MCPCatalogItem` | Get a single catalog item |
| `ListConnections(ctx, *ConnectionParams)` | `*MCPConnectionListResponse` | List end-user connections |
| `DeleteConnection(ctx, id)` | `error` | Delete an end-user connection |

### Standalone functions

| Function | Returns | Description |
|----------|---------|-------------|
| `GenerateConnectionToken(serverID, ownerID, endUserID, key)` | `(string, error)` | Create a signed JWT (1 hr expiry) |
| `GenerateConnectionLink(baseURL, serverID, ownerID, endUserID, key)` | `(string, error)` | Build a full connection URL |

## Requirements

- Go ≥ 1.21
- No external dependencies (stdlib only)

## License

MIT
