package splox

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// WorkflowService provides methods for the Workflows API.
type WorkflowService struct {
	client *Client
}

// ListParams are optional parameters for [WorkflowService.List].
type ListParams struct {
	Limit  int
	Cursor string
	Search string
}

// List returns the authenticated user's workflows.
func (s *WorkflowService) List(ctx context.Context, params *ListParams) (*WorkflowListResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Cursor != "" {
			v.Set("cursor", params.Cursor)
		}
		if params.Search != "" {
			v.Set("search", params.Search)
		}
	}

	var resp WorkflowListResponse
	if err := s.client.do(ctx, "GET", addParams("/workflows", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get returns a workflow with its draft version, nodes, and edges.
func (s *WorkflowService) Get(ctx context.Context, workflowID string) (*WorkflowFullResponse, error) {
	var resp WorkflowFullResponse
	if err := s.client.do(ctx, "GET", "/workflows/"+workflowID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLatestVersion returns the latest version of a workflow.
func (s *WorkflowService) GetLatestVersion(ctx context.Context, workflowID string) (*WorkflowVersion, error) {
	var resp WorkflowVersion
	if err := s.client.do(ctx, "GET", "/workflows/"+workflowID+"/versions/latest", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVersions returns all versions of a workflow.
func (s *WorkflowService) ListVersions(ctx context.Context, workflowID string) (*WorkflowVersionListResponse, error) {
	var resp WorkflowVersionListResponse
	if err := s.client.do(ctx, "GET", "/workflows/"+workflowID+"/versions", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStartNodes returns the start nodes for a workflow version.
func (s *WorkflowService) GetStartNodes(ctx context.Context, workflowVersionID string) (*StartNodesResponse, error) {
	var resp StartNodesResponse
	if err := s.client.do(ctx, "GET", "/workflows/"+workflowVersionID+"/start-nodes", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RunParams are the parameters for [WorkflowService.Run].
type RunParams struct {
	WorkflowVersionID string                `json:"workflow_version_id"`
	ChatID            string                `json:"chat_id"`
	StartNodeID       string                `json:"start_node_id"`
	Query             string                `json:"query"`
	Files             []WorkflowRequestFile `json:"files,omitempty"`
	AdditionalParams  map[string]any        `json:"additional_params,omitempty"`
}

// Run triggers a workflow execution.
func (s *WorkflowService) Run(ctx context.Context, params RunParams) (*RunResponse, error) {
	var resp RunResponse
	if err := s.client.do(ctx, "POST", "/workflow-requests/run", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Listen opens an SSE stream for real-time execution updates.
// The caller must call [SSEIter.Close] when done.
func (s *WorkflowService) Listen(ctx context.Context, workflowRequestID string) (*SSEIter, error) {
	return s.client.streamSSE(ctx, "/workflow-requests/"+workflowRequestID+"/listen")
}

// GetExecutionTree returns the complete execution hierarchy.
func (s *WorkflowService) GetExecutionTree(ctx context.Context, workflowRequestID string) (*ExecutionTreeResponse, error) {
	var resp ExecutionTreeResponse
	if err := s.client.do(ctx, "GET", "/workflow-requests/"+workflowRequestID+"/execution-tree", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// HistoryParams are optional parameters for [WorkflowService.GetHistory].
type HistoryParams struct {
	Limit  int
	Cursor string
	Search string
}

// GetHistory returns paginated execution history.
func (s *WorkflowService) GetHistory(ctx context.Context, workflowRequestID string, params *HistoryParams) (*HistoryResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Cursor != "" {
			v.Set("cursor", params.Cursor)
		}
		if params.Search != "" {
			v.Set("search", params.Search)
		}
	}

	var resp HistoryResponse
	if err := s.client.do(ctx, "GET", addParams("/workflow-requests/"+workflowRequestID+"/history", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Stop cancels a running workflow execution.
func (s *WorkflowService) Stop(ctx context.Context, workflowRequestID string) error {
	return s.client.do(ctx, "POST", "/workflow-requests/"+workflowRequestID+"/stop", nil, nil)
}

// RunAndWait triggers a workflow and blocks until it reaches a terminal state.
// It returns the full execution tree on completion.
func (s *WorkflowService) RunAndWait(ctx context.Context, params RunParams, timeout time.Duration) (*ExecutionTreeResponse, error) {
	result, err := s.Run(ctx, params)
	if err != nil {
		return nil, err
	}

	// Create a context with timeout for the SSE wait
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	iter, err := s.Listen(waitCtx, result.WorkflowRequestID)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	terminal := map[string]bool{
		"completed": true,
		"failed":    true,
		"stopped":   true,
	}

	for iter.Next() {
		ev := iter.Event()
		if ev.WorkflowRequest != nil && terminal[ev.WorkflowRequest.Status] {
			return s.GetExecutionTree(ctx, result.WorkflowRequestID)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	// Check if context timed out
	if waitCtx.Err() != nil {
		return nil, &TimeoutError{Message: fmt.Sprintf("workflow did not complete within %s", timeout)}
	}

	// Stream ended without terminal status — fetch tree anyway
	return s.GetExecutionTree(ctx, result.WorkflowRequestID)
}
