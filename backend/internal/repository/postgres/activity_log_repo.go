package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

type ActivityLogRepo struct {
	db *sqlx.DB
}

func NewActivityLogRepo(db *sqlx.DB) *ActivityLogRepo {
	return &ActivityLogRepo{db: db}
}

func (r *ActivityLogRepo) Create(ctx context.Context, log *domain.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (
			id, user_id, store_id, action_type, module, reference_id, metadata, created_at
		) VALUES (
			uuid_generate_v4(), :user_id, :store_id, :action_type, :module, :reference_id, :metadata, NOW()
		)`

	_, err := r.db.NamedExecContext(ctx, query, log)
	if err != nil {
		return fmt.Errorf("inserting activity log: %w", err)
	}
	return nil
}

func (r *ActivityLogRepo) FindAll(ctx context.Context, storeID string, filter dto.ActivityLogFilter) ([]*domain.ActivityLog, int, error) {
	where := []string{"1=1"}
	args := map[string]interface{}{}

	// If storeID is provided, filter by it. BUT global auth events might have NULL store_id.
	// However, usually we view logs within a store context.
	if storeID != "" {
		where = append(where, "(store_id = :store_id OR store_id IS NULL)")
		args["store_id"] = storeID
	}

	if filter.UserID != "" {
		where = append(where, "user_id = :user_id")
		args["user_id"] = filter.UserID
	}
	if filter.Module != "" {
		where = append(where, "module = :module")
		args["module"] = filter.Module
	}
	if filter.ActionType != "" {
		where = append(where, "action_type = :action_type")
		args["action_type"] = filter.ActionType
	}
	if filter.StartDate != "" {
		where = append(where, "created_at >= :start_date")
		args["start_date"] = filter.StartDate
	}
	if filter.EndDate != "" {
		where = append(where, "created_at <= :end_date")
		args["end_date"] = filter.EndDate
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activity_logs WHERE %s", whereClause)
	nstmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("preparing count query: %w", err)
	}
	if err := nstmt.GetContext(ctx, &total, args); err != nil {
		return nil, 0, fmt.Errorf("executing count query: %w", err)
	}

	// Fetch page
	query := fmt.Sprintf(`
		SELECT 
			al.*, 
			u.name as user_name
		FROM activity_logs al
		JOIN users u ON al.user_id = u.id
		WHERE %s
		ORDER BY al.created_at DESC
		LIMIT :limit OFFSET :offset`, whereClause)

	args["limit"] = filter.PerPage
	args["offset"] = filter.Offset()

	logs := make([]*domain.ActivityLog, 0)
	nstmt, err = r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("preparing fetch query: %w", err)
	}
	if err := nstmt.SelectContext(ctx, &logs, args); err != nil {
		return nil, 0, fmt.Errorf("executing fetch query: %w", err)
	}

	return logs, total, nil
}
