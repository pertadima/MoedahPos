package dto

import "github.com/moedahpos/backend/internal/domain"

type ActivityLogResponse struct {
	ID          string               `json:"id"`
	UserID      string               `json:"user_id"`
	UserName    string               `json:"user_name"`
	StoreID     *string              `json:"store_id"`
	ActionType  domain.ActionType    `json:"action_type"`
	Module      domain.ActivityModule `json:"module"`
	ReferenceID *string              `json:"reference_id"`
	Metadata    interface{}          `json:"metadata"`
	CreatedAt   string               `json:"created_at"`
}

type ActivityLogFilter struct {
	PaginationQuery
	StoreID    string `json:"store_id"`
	UserID     string `json:"user_id"`
	Module     string `json:"module"`
	ActionType string `json:"action_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}
