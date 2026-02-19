package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// SubscriptionRepo implements identity.SubscriptionRepository.
type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

// NewSubscriptionRepo creates a new SubscriptionRepo.
func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *identity.Subscription) error {
	s.ID = types.NewID()
	query := `
		INSERT INTO subscriptions (id, tenant_id, plan, billing_start, billing_end, auto_revoke, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		s.ID, s.TenantID, s.Plan, s.BillingStart, s.BillingEnd, s.AutoRevoke, s.Status,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *SubscriptionRepo) GetByTenantID(ctx context.Context, tenantID types.ID) (*identity.Subscription, error) {
	query := `
		SELECT id, tenant_id, plan, billing_start, billing_end, auto_revoke, status, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1`

	s := &identity.Subscription{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&s.ID, &s.TenantID, &s.Plan, &s.BillingStart, &s.BillingEnd,
		&s.AutoRevoke, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Subscription", tenantID)
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepo) GetAllActive(ctx context.Context) ([]identity.Subscription, error) {
	query := `
		SELECT id, tenant_id, plan, billing_start, billing_end, auto_revoke, status, created_at, updated_at
		FROM subscriptions
		WHERE status = 'ACTIVE'`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query active subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []identity.Subscription
	for rows.Next() {
		var s identity.Subscription
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.Plan, &s.BillingStart, &s.BillingEnd,
			&s.AutoRevoke, &s.Status, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, s *identity.Subscription) error {
	query := `
		UPDATE subscriptions
		SET plan = $2, billing_start = $3, billing_end = $4, auto_revoke = $5,
		    status = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		s.ID, s.Plan, s.BillingStart, s.BillingEnd, s.AutoRevoke, s.Status,
	).Scan(&s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("Subscription", s.ID)
		}
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

var _ identity.SubscriptionRepository = (*SubscriptionRepo)(nil)
