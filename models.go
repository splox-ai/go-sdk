package splox

// WorkflowRequestFile represents a file attached to a workflow run request.
type WorkflowRequestFile struct {
	URL         string         `json:"url"`
	ContentType string         `json:"content_type,omitempty"`
	FileName    string         `json:"file_name,omitempty"`
	FileSize    int64          `json:"file_size,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// --- Workflow / Version / Node / Edge ---

type Workflow struct {
	ID            string           `json:"id"`
	UserID        string           `json:"user_id"`
	CreatedAt     string           `json:"created_at,omitempty"`
	UpdatedAt     string           `json:"updated_at,omitempty"`
	LatestVersion *WorkflowVersion `json:"latest_version,omitempty"`
	IsPublic      *bool            `json:"is_public,omitempty"`
}

type WorkflowVersion struct {
	ID            string         `json:"id"`
	WorkflowID    string         `json:"workflow_id"`
	VersionNumber int            `json:"version_number"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	Description   string         `json:"description,omitempty"`
	CreatedAt     string         `json:"created_at,omitempty"`
	UpdatedAt     string         `json:"updated_at,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Node struct {
	ID                string         `json:"id"`
	WorkflowVersionID string         `json:"workflow_version_id"`
	NodeType          string         `json:"node_type"`
	Label             string         `json:"label"`
	PosX              *float64       `json:"pos_x,omitempty"`
	PosY              *float64       `json:"pos_y,omitempty"`
	ParentID          string         `json:"parent_id,omitempty"`
	Extent            string         `json:"extent,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
}

type Edge struct {
	ID                string         `json:"id"`
	WorkflowVersionID string         `json:"workflow_version_id"`
	Source            string         `json:"source"`
	Target            string         `json:"target"`
	EdgeType          string         `json:"edge_type"`
	SourceHandle      string         `json:"source_handle,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
}

// --- Execution ---

type WorkflowRequest struct {
	ID                        string         `json:"id"`
	WorkflowVersionID         string         `json:"workflow_version_id"`
	StartNodeID               string         `json:"start_node_id"`
	Status                    string         `json:"status"`
	CreatedAt                 string         `json:"created_at"`
	UserID                    string         `json:"user_id,omitempty"`
	BillingUserID             string         `json:"billing_user_id,omitempty"`
	ParentNodeExecutionID     string         `json:"parent_node_execution_id,omitempty"`
	ParentWorkflowRequestID   string         `json:"parent_workflow_request_id,omitempty"`
	ChatID                    string         `json:"chat_id,omitempty"`
	Payload                   map[string]any `json:"payload,omitempty"`
	Metadata                  map[string]any `json:"metadata,omitempty"`
	StartedAt                 string         `json:"started_at,omitempty"`
	CompletedAt               string         `json:"completed_at,omitempty"`
}

type NodeExecution struct {
	ID                string         `json:"id"`
	WorkflowRequestID string         `json:"workflow_request_id"`
	NodeID            string         `json:"node_id"`
	WorkflowVersionID string         `json:"workflow_version_id"`
	Status            string         `json:"status"`
	InputData         map[string]any `json:"input_data,omitempty"`
	OutputData        map[string]any `json:"output_data,omitempty"`
	AttemptCount      *int           `json:"attempt_count,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	CompletedAt       string         `json:"completed_at,omitempty"`
	FailedAt          string         `json:"failed_at,omitempty"`
}

type ChildExecution struct {
	Index             int             `json:"index"`
	WorkflowRequestID string          `json:"workflow_request_id"`
	Status            string          `json:"status"`
	Label             string          `json:"label,omitempty"`
	TargetNodeLabel   string          `json:"target_node_label,omitempty"`
	CreatedAt         string          `json:"created_at,omitempty"`
	CompletedAt       string          `json:"completed_at,omitempty"`
	Nodes             []ExecutionNode `json:"nodes,omitempty"`
}

type ExecutionNode struct {
	ID               string           `json:"id"`
	NodeID           string           `json:"node_id"`
	Status           string           `json:"status"`
	NodeLabel        string           `json:"node_label,omitempty"`
	NodeType         string           `json:"node_type,omitempty"`
	InputData        map[string]any   `json:"input_data,omitempty"`
	OutputData       map[string]any   `json:"output_data,omitempty"`
	CreatedAt        string           `json:"created_at,omitempty"`
	CompletedAt      string           `json:"completed_at,omitempty"`
	FailedAt         string           `json:"failed_at,omitempty"`
	AttemptCount     *int             `json:"attempt_count,omitempty"`
	ChildExecutions  []ChildExecution `json:"child_executions,omitempty"`
	TotalChildren    *int             `json:"total_children,omitempty"`
	HasMoreChildren  *bool            `json:"has_more_children,omitempty"`
}

type ExecutionTree struct {
	WorkflowRequestID string          `json:"workflow_request_id"`
	Status            string          `json:"status"`
	CreatedAt         string          `json:"created_at"`
	CompletedAt       string          `json:"completed_at,omitempty"`
	Nodes             []ExecutionNode `json:"nodes,omitempty"`
}

// --- Chat ---

type Chat struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	UserID           string         `json:"user_id,omitempty"`
	ResourceType     string         `json:"resource_type,omitempty"`
	ResourceID       string         `json:"resource_id,omitempty"`
	IsPublic         *bool          `json:"is_public,omitempty"`
	PublicShareToken string         `json:"public_share_token,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CreatedAt        string         `json:"created_at,omitempty"`
	UpdatedAt        string         `json:"updated_at,omitempty"`
}

type ChatMessageContent struct {
	Type       string         `json:"type"`
	Text       string         `json:"text,omitempty"`
	ToolCallID string         `json:"toolCallId,omitempty"`
	ToolName   string         `json:"toolName,omitempty"`
	Args       map[string]any `json:"args,omitempty"`
	Result     any            `json:"result,omitempty"`
	Reasoning  string         `json:"reasoning,omitempty"`
}

type ChatMessage struct {
	ID        string               `json:"id"`
	ChatID    string               `json:"chat_id"`
	Role      string               `json:"role"`
	Content   []ChatMessageContent `json:"content,omitempty"`
	ParentID  string               `json:"parent_id,omitempty"`
	Status    map[string]any       `json:"status,omitempty"`
	Metadata  map[string]any       `json:"metadata,omitempty"`
	Files     []map[string]any     `json:"files,omitempty"`
	CreatedAt string               `json:"created_at,omitempty"`
	UpdatedAt string               `json:"updated_at,omitempty"`
}

// --- Pagination ---

type Pagination struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// --- SSE ---

// SSEEvent represents a Server-Sent Event from a listen stream.
// For chat events, check EventType:
//   - "text_delta": TextDelta contains the streamed text chunk
//   - "reasoning_delta": ReasoningDelta contains thinking text
//   - "tool_call_start": Tool call initiated (ToolCallID, ToolName)
//   - "tool_call_delta": Tool call args delta (ToolCallID, ToolArgsDelta)
//   - "tool_start": Tool execution started (ToolName, ToolCallID)
//   - "tool_complete": Tool finished (ToolName, ToolCallID, ToolResult)
//   - "tool_error": Tool failed (ToolName, ToolCallID, Error)
//   - "tool_approval_request": Approval needed (ToolName, ToolCallID, ToolArgs)
//   - "tool_approval_response": Approval result (ToolName, ToolCallID, Approved)
//   - "user_message": Voice transcript (Text)
//   - "done": Iteration complete
//   - "stopped": User stopped workflow
//   - "error": Error occurred (Error)
type SSEEvent struct {
	WorkflowRequest *WorkflowRequest `json:"workflow_request,omitempty"`
	NodeExecution   *NodeExecution   `json:"node_execution,omitempty"`
	IsKeepalive     bool             `json:"-"`
	RawData         string           `json:"-"`

	// Event type and metadata
	EventType string `json:"type,omitempty"`
	Iteration *int   `json:"iteration,omitempty"`
	RunID     string `json:"run_id,omitempty"`

	// Text streaming
	TextDelta string `json:"delta,omitempty"`

	// Reasoning/thinking
	ReasoningDelta string `json:"reasoning_delta,omitempty"`
	ReasoningType  string `json:"reasoning_type,omitempty"`

	// Tool calls
	ToolCallID    string `json:"tool_call_id,omitempty"`
	ToolName      string `json:"tool_name,omitempty"`
	ToolArgsDelta string `json:"tool_args_delta,omitempty"`
	ToolArgs      any    `json:"args,omitempty"`
	ToolResult    any    `json:"result,omitempty"`

	// Tool approval
	Approved *bool `json:"approved,omitempty"`

	// Messages and errors
	Text    string `json:"text,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// --- Response types ---

type WorkflowListResponse struct {
	Workflows  []Workflow `json:"workflows"`
	Pagination Pagination `json:"pagination"`
}

type WorkflowFullResponse struct {
	Workflow        Workflow        `json:"workflow"`
	WorkflowVersion WorkflowVersion `json:"workflow_version"`
	Nodes           []Node          `json:"nodes"`
	Edges           []Edge          `json:"edges"`
}

type StartNodesResponse struct {
	Nodes []Node `json:"nodes"`
}

type WorkflowVersionListResponse struct {
	Versions []WorkflowVersion `json:"versions"`
}

type RunResponse struct {
	WorkflowRequestID string `json:"workflow_request_id"`
}

type ExecutionTreeResponse struct {
	ExecutionTree ExecutionTree `json:"execution_tree"`
}

type HistoryResponse struct {
	Data       []WorkflowRequest `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

type ChatListResponse struct {
	Chats []Chat `json:"chats"`
}

type ChatHistoryResponse struct {
	Messages []ChatMessage `json:"messages"`
	HasMore  bool          `json:"has_more"`
}

type EventResponse struct {
	OK      bool   `json:"ok"`
	EventID string `json:"event_id"`
}

// --- Billing / Cost Tracking ---

type UserBalance struct {
	BalanceMicrodollars int64   `json:"balance_microdollars"`
	BalanceUSD          float64 `json:"balance_usd"`
	Currency            string  `json:"currency"`
}

type BalanceTransaction struct {
	ID                    string          `json:"id"`
	UserID                string          `json:"user_id"`
	Amount                int64           `json:"amount"`
	Currency              string          `json:"currency"`
	Type                  string          `json:"type"`
	Status                string          `json:"status"`
	Description           *string         `json:"description,omitempty"`
	Metadata              map[string]any  `json:"metadata,omitempty"`
	StripePaymentIntentID *string         `json:"stripe_payment_intent_id,omitempty"`
	StripeChargeID        *string         `json:"stripe_charge_id,omitempty"`
	CreatedAt             string          `json:"created_at"`
	UpdatedAt             string          `json:"updated_at"`
}

type TransactionPagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalCount int  `json:"total_count"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type TransactionHistoryResponse struct {
	Transactions []BalanceTransaction  `json:"transactions"`
	Pagination   TransactionPagination `json:"pagination"`
}

type ActivityStats struct {
	Balance           float64 `json:"balance"`
	TotalRequests     int     `json:"total_requests"`
	TotalSpending     float64 `json:"total_spending"`
	AvgCostPerRequest float64 `json:"avg_cost_per_request"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
}

type DailyActivity struct {
	Date         string  `json:"date"`
	TotalCost    float64 `json:"total_cost"`
	RequestCount int     `json:"request_count"`
	NodeCount    int     `json:"node_count"`
}

type DailyActivityResponse struct {
	Data []DailyActivity `json:"data"`
	Days int             `json:"days"`
}
