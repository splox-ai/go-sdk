package splox

import (
	"context"
	"fmt"
	"net/url"
)

// BillingService provides methods for balance, cost tracking, and activity.
type BillingService struct {
	client *Client
}

// GetBalance returns the authenticated user's current balance.
func (s *BillingService) GetBalance(ctx context.Context) (*UserBalance, error) {
	var resp UserBalance
	if err := s.client.do(ctx, "GET", "/billing/balance", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TransactionHistoryParams are optional filters for [BillingService.GetTransactionHistory].
type TransactionHistoryParams struct {
	Page      int
	Limit     int
	Types     string // comma-separated: "credit", "debit", "refund"
	Statuses  string // comma-separated: "pending", "completed", "failed"
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	MinAmount float64
	MaxAmount float64
	Search    string
}

// GetTransactionHistory returns paginated, filterable transaction history.
func (s *BillingService) GetTransactionHistory(ctx context.Context, params *TransactionHistoryParams) (*TransactionHistoryResponse, error) {
	v := url.Values{}
	if params != nil {
		if params.Page > 0 {
			v.Set("page", fmt.Sprintf("%d", params.Page))
		}
		if params.Limit > 0 {
			v.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Types != "" {
			v.Set("types", params.Types)
		}
		if params.Statuses != "" {
			v.Set("statuses", params.Statuses)
		}
		if params.StartDate != "" {
			v.Set("start_date", params.StartDate)
		}
		if params.EndDate != "" {
			v.Set("end_date", params.EndDate)
		}
		if params.MinAmount > 0 {
			v.Set("min_amount", fmt.Sprintf("%f", params.MinAmount))
		}
		if params.MaxAmount > 0 {
			v.Set("max_amount", fmt.Sprintf("%f", params.MaxAmount))
		}
		if params.Search != "" {
			v.Set("search", params.Search)
		}
	}

	var resp TransactionHistoryResponse
	if err := s.client.do(ctx, "GET", addParams("/billing/transactions", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetActivityStats returns aggregate activity statistics (balance, total
// requests, total spending, average cost per request, and token counts).
func (s *BillingService) GetActivityStats(ctx context.Context) (*ActivityStats, error) {
	var resp ActivityStats
	if err := s.client.do(ctx, "GET", "/activity/stats", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DailyActivityParams are optional parameters for [BillingService.GetDailyActivity].
type DailyActivityParams struct {
	Days int // number of days to look back (default 30)
}

// GetDailyActivity returns daily aggregated spending and usage data.
func (s *BillingService) GetDailyActivity(ctx context.Context, params *DailyActivityParams) (*DailyActivityResponse, error) {
	v := url.Values{}
	if params != nil && params.Days > 0 {
		v.Set("days", fmt.Sprintf("%d", params.Days))
	}

	var resp DailyActivityResponse
	if err := s.client.do(ctx, "GET", addParams("/activity/daily", v), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
