package splox

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// MCPService provides methods for MCP catalog browsing and connection management.
type MCPService struct {
	client *Client
}

// --------------------------------------------------------------------------
// Catalog
// --------------------------------------------------------------------------

// CatalogParams are optional filters for [MCPService.ListCatalog].
type CatalogParams struct {
	Page     int
	PerPage  int
	Search   string
	Featured bool
}

// ListCatalog returns a paginated list of MCP servers from the catalog.
func (s *MCPService) ListCatalog(ctx context.Context, params *CatalogParams) (*MCPCatalogListResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Page > 0 {
			v.Set("page", fmt.Sprintf("%d", params.Page))
		}
		if params.PerPage > 0 {
			v.Set("per_page", fmt.Sprintf("%d", params.PerPage))
		}
		if params.Search != "" {
			v.Set("search", params.Search)
		}
		if params.Featured {
			v.Set("featured", "true")
		}
	}

	var resp MCPCatalogListResponse
	if err := s.client.do(ctx, "GET", addParams("/mcp-catalog", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCatalogItem returns a single MCP server from the catalog by ID.
func (s *MCPService) GetCatalogItem(ctx context.Context, id string) (*MCPCatalogItem, error) {
	var resp MCPCatalogResponse
	if err := s.client.do(ctx, "GET", "/mcp-catalog/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return resp.MCPServer, nil
}

// --------------------------------------------------------------------------
// Connections
// --------------------------------------------------------------------------

// ConnectionParams are optional filters for [MCPService.ListConnections].
type ConnectionParams struct {
	MCPServerID string
	EndUserID   string
}

// ListConnections returns MCP connections for the authenticated user.
func (s *MCPService) ListConnections(ctx context.Context, params *ConnectionParams) (*MCPConnectionListResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.MCPServerID != "" {
			v.Set("mcp_server_id", params.MCPServerID)
		}
		if params.EndUserID != "" {
			v.Set("end_user_id", params.EndUserID)
		}
	}

	var resp MCPConnectionListResponse
	if err := s.client.do(ctx, "GET", addParams("/mcp-connections", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteConnection deletes an end-user MCP connection by ID.
func (s *MCPService) DeleteConnection(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/mcp-connections/"+id, nil, nil)
}

// ExecuteToolParams are parameters for [MCPService.ExecuteTool].
type ExecuteToolParams struct {
	MCPServerID string         `json:"mcp_server_id"`
	ToolSlug    string         `json:"tool_slug"`
	Args        map[string]any `json:"args,omitempty"`
}

// ExecuteTool executes a tool on a caller-owned MCP server.
func (s *MCPService) ExecuteTool(ctx context.Context, params ExecuteToolParams) (*MCPExecuteToolResponse, error) {
	body := params
	if body.Args == nil {
		body.Args = map[string]any{}
	}

	var resp MCPExecuteToolResponse
	if err := s.client.do(ctx, "POST", "/mcp-tools/execute", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListUserServers lists caller-owned MCP servers.
func (s *MCPService) ListUserServers(ctx context.Context) (*UserMCPServerListResponse, error) {
	var resp UserMCPServerListResponse
	if err := s.client.do(ctx, "GET", "/user-mcp-servers", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetServerTools lists tools for a caller-owned MCP server.
func (s *MCPService) GetServerTools(ctx context.Context, mcpServerID string) (*MCPServerToolsResponse, error) {
	var resp MCPServerToolsResponse
	if err := s.client.do(ctx, "GET", "/user-mcp-servers/"+mcpServerID+"/tools", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --------------------------------------------------------------------------
// Connection Token (client-side JWT generation)
// --------------------------------------------------------------------------

const (
	mcpConnectionIssuer = "splox-mcp-connection"
	mcpConnectionExpiry = 1 * time.Hour
)

// deriveSigningKey derives a JWT HMAC-SHA256 signing key from the credentials
// encryption key, matching the backend's derivation scheme.
func deriveSigningKey(credentialsEncryptionKey string) []byte {
	h := sha256.New()
	h.Write([]byte("mcp-connection-jwt:" + credentialsEncryptionKey))
	return h.Sum(nil)
}

// base64URLEncode encodes bytes as unpadded base64url.
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// GenerateConnectionToken creates a signed JWT for end-user credential
// submission. The token embeds the MCP server ID, owner user ID, and end-user
// ID. It expires after 1 hour.
//
// This is equivalent to the backend's mcp.GenerateConnectionToken and lets SDK
// consumers generate tokens without a round-trip to the API.
func GenerateConnectionToken(mcpServerID, ownerUserID, endUserID, credentialsEncryptionKey string) (string, error) {
	now := time.Now().UTC()

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := map[string]interface{}{
		"mcp_server_id": mcpServerID,
		"owner_user_id": ownerUserID,
		"end_user_id":   endUserID,
		"iss":           mcpConnectionIssuer,
		"iat":           now.Unix(),
		"exp":           now.Add(mcpConnectionExpiry).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("splox: marshal JWT header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("splox: marshal JWT claims: %w", err)
	}

	signingInput := base64URLEncode(headerJSON) + "." + base64URLEncode(claimsJSON)

	mac := hmac.New(sha256.New, deriveSigningKey(credentialsEncryptionKey))
	mac.Write([]byte(signingInput))
	signature := base64URLEncode(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// GenerateConnectionLink builds a full connection URL that end-users can visit
// to submit their credentials for a specific MCP server.
//
// baseURL is the Splox application URL (e.g. "https://app.splox.io").
func GenerateConnectionLink(baseURL, mcpServerID, ownerUserID, endUserID, credentialsEncryptionKey string) (string, error) {
	token, err := GenerateConnectionToken(mcpServerID, ownerUserID, endUserID, credentialsEncryptionKey)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/tools/connect?token=%s", strings.TrimRight(baseURL, "/"), token), nil
}
