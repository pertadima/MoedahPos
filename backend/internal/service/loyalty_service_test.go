package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// ─── Mocks ────────────────────────────────────────────────────────────────────

// mockLoyaltyRepo implements repository.LoyaltyRepository.
type mockLoyaltyRepo struct {
	balance    float64
	balanceErr error
	earnEntry  *domain.LoyaltyLedger
	earnErr    error
	spendEntry *domain.LoyaltyLedger
	spendErr   error
	history    []*domain.LoyaltyLedger
	historyErr error
	assignErr  error
	tier       *domain.MembershipTier
	tierErr    error
}

func (m *mockLoyaltyRepo) GetBalance(_ context.Context, _ string) (float64, error) {
	return m.balance, m.balanceErr
}
func (m *mockLoyaltyRepo) EarnPoints(_ context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	if m.earnErr != nil {
		return nil, m.earnErr
	}
	if m.earnEntry != nil {
		return m.earnEntry, nil
	}
	return &domain.LoyaltyLedger{
		ID:            "ledger-1",
		CustomerID:    customerID,
		PointsDelta:   points,
		TransactionID: transactionID,
		Type:          "EARN",
		CreatedAt:     time.Now(),
	}, nil
}
func (m *mockLoyaltyRepo) SpendPoints(_ context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	if m.spendErr != nil {
		return nil, m.spendErr
	}
	if m.spendEntry != nil {
		return m.spendEntry, nil
	}
	return &domain.LoyaltyLedger{
		ID:            "ledger-2",
		CustomerID:    customerID,
		PointsDelta:   -points,
		TransactionID: transactionID,
		Type:          "SPEND",
		CreatedAt:     time.Now(),
	}, nil
}
func (m *mockLoyaltyRepo) GetHistory(_ context.Context, _ string) ([]*domain.LoyaltyLedger, error) {
	return m.history, m.historyErr
}
func (m *mockLoyaltyRepo) AssignTier(_ context.Context, _, _ string) error { return m.assignErr }
func (m *mockLoyaltyRepo) GetCustomerTier(_ context.Context, _ string) (*domain.MembershipTier, error) {
	return m.tier, m.tierErr
}

// mockLoyaltyRepoWithEarnTracker wraps mockLoyaltyRepo with an earnCalled callback
// so we can detect if EarnPoints is called even when points are zero.
type mockLoyaltyRepoWithEarnTracker struct {
	mockLoyaltyRepo
	onEarnCalled func()
}

func (m *mockLoyaltyRepoWithEarnTracker) EarnPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	if m.onEarnCalled != nil {
		m.onEarnCalled()
	}
	return m.mockLoyaltyRepo.EarnPoints(ctx, customerID, transactionID, points)
}

// mockTierRepo implements repository.MembershipTierRepository.
type mockTierRepo struct {
	tiers    []*domain.MembershipTier
	findErr  error
	tier     *domain.MembershipTier
	findByID *domain.MembershipTier
	byIDErr  error
}

func (m *mockTierRepo) FindAll(_ context.Context) ([]*domain.MembershipTier, error) {
	return m.tiers, m.findErr
}
func (m *mockTierRepo) FindByID(_ context.Context, _ string) (*domain.MembershipTier, error) {
	return m.findByID, m.byIDErr
}

// mockCustomerRepoLoyalty implements repository.CustomerRepository (minimal stubs).
type mockCustomerRepoLoyalty struct{}

func (m *mockCustomerRepoLoyalty) Create(_ context.Context, c *domain.Customer) (*domain.Customer, error) {
	return c, nil
}
func (m *mockCustomerRepoLoyalty) FindAll(_ context.Context, _ dto.CustomerListFilter) ([]*domain.Customer, int, error) {
	return nil, 0, nil
}
func (m *mockCustomerRepoLoyalty) FindByID(_ context.Context, _ string) (*domain.Customer, error) {
	return nil, nil
}
func (m *mockCustomerRepoLoyalty) Update(_ context.Context, c *domain.Customer) (*domain.Customer, error) {
	return c, nil
}
func (m *mockCustomerRepoLoyalty) SoftDelete(_ context.Context, _ string) error { return nil }
func (m *mockCustomerRepoLoyalty) SearchByPhone(_ context.Context, _, _ string) ([]*domain.Customer, error) {
	return nil, nil
}
func (m *mockCustomerRepoLoyalty) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.Customer, error) {
	return nil, nil
}

// Ensure mocks satisfy interfaces.
var _ repository.LoyaltyRepository = (*mockLoyaltyRepo)(nil)
var _ repository.MembershipTierRepository = (*mockTierRepo)(nil)
var _ repository.CustomerRepository = (*mockCustomerRepoLoyalty)(nil)

// ─── CalculatePoints — pure function tests ────────────────────────────────────

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		name       string
		total      float64
		multiplier float64
		want       float64
	}{
		{
			name:       "Bronze tier — exact boundary",
			total:      50000,
			multiplier: 1.0,
			want:       50, // floor(50000/1000)*1.0
		},
		{
			name:       "Silver tier — fractional total floored",
			total:      51999,
			multiplier: 1.5,
			want:       76.5, // floor(51999/1000)=51, 51*1.5=76.5
		},
		{
			name:       "Gold tier — whole number",
			total:      100000,
			multiplier: 2.0,
			want:       200,
		},
		{
			name:       "Platinum tier",
			total:      30000,
			multiplier: 3.0,
			want:       90,
		},
		{
			name:       "Zero total returns zero points",
			total:      0,
			multiplier: 2.0,
			want:       0,
		},
		{
			name:       "Negative total returns zero points",
			total:      -1000,
			multiplier: 1.0,
			want:       0,
		},
		{
			name:       "Zero multiplier returns zero points",
			total:      50000,
			multiplier: 0,
			want:       0,
		},
		{
			name:       "Total less than one unit — no points",
			total:      999,
			multiplier: 1.0,
			want:       0, // floor(999/1000)=0
		},
		{
			name:       "Exactly one unit — one point",
			total:      1000,
			multiplier: 1.0,
			want:       1,
		},
		{
			name:       "Future-dated timestamp scenario — calculation identical",
			total:      75000, // timestamp doesn't affect CalculatePoints
			multiplier: 1.5,
			want:       112.5, // floor(75000/1000)*1.5 = 75*1.5
		},
		{
			name:       "Very large total",
			total:      10_000_000,
			multiplier: 2.0,
			want:       20_000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculatePoints(tc.total, tc.multiplier, 1000)
			if got != tc.want {
				t.Errorf("CalculatePoints(%.2f, %.2f) = %.2f; want %.2f", tc.total, tc.multiplier, got, tc.want)
			}
		})
	}
}

// ─── ListTiers ────────────────────────────────────────────────────────────────

func TestLoyaltyService_ListTiers(t *testing.T) {
	nop := zerolog.Nop()

	t.Run("returns all tiers", func(t *testing.T) {
		tierRepo := &mockTierRepo{tiers: []*domain.MembershipTier{
			{ID: "t1", Name: "Bronze", Multiplier: 1.0},
			{ID: "t2", Name: "Gold", Multiplier: 2.0},
		}}
		svc := NewLoyaltyService(&mockLoyaltyRepo{}, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.ListTiers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("want 2 tiers, got %d", len(resp))
		}
		if resp[0].Name != "Bronze" || resp[1].Name != "Gold" {
			t.Error("tier names mismatch")
		}
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		tierRepo := &mockTierRepo{findErr: errors.New("db down")}
		svc := NewLoyaltyService(&mockLoyaltyRepo{}, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.ListTiers(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// ─── GetBalance ───────────────────────────────────────────────────────────────

func TestLoyaltyService_GetBalance(t *testing.T) {
	nop := zerolog.Nop()

	t.Run("returns balance with tier info", func(t *testing.T) {
		lr := &mockLoyaltyRepo{
			balance: 250,
			tier:    &domain.MembershipTier{ID: "t1", Name: "Gold", Multiplier: 2.0},
		}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.GetBalance(context.Background(), "cust-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Balance != 250 {
			t.Errorf("want balance 250, got %.2f", resp.Balance)
		}
		if resp.Tier == nil || resp.Tier.Name != "Gold" {
			t.Error("expected Gold tier")
		}
	})

	t.Run("returns balance with no tier assigned", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 100, tier: nil}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.GetBalance(context.Background(), "cust-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Tier != nil {
			t.Error("expected nil tier")
		}
	})

	t.Run("balance repo error propagated", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balanceErr: errors.New("db failure")}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.GetBalance(context.Background(), "cust-3")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("tier fetch error propagated", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 50, tierErr: errors.New("tier db error")}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.GetBalance(context.Background(), "cust-4")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ─── EarnPoints ───────────────────────────────────────────────────────────────

//nolint:gocognit,funlen // test function with many subtests is inherently long
func TestLoyaltyService_EarnPoints(t *testing.T) {
	nop := zerolog.Nop()
	txID := "txn-abc"

	t.Run("earns points at Bronze tier (1x)", func(t *testing.T) {
		lr := &mockLoyaltyRepo{tier: &domain.MembershipTier{Multiplier: 1.0}}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.EarnPoints(context.Background(), "", "c1", &txID, 50000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PointsDelta != 50 {
			t.Errorf("want 50 points, got %.2f", resp.PointsDelta)
		}
		if resp.Type != "EARN" {
			t.Error("expected EARN type")
		}
	})

	t.Run("earns points at Gold tier (2x)", func(t *testing.T) {
		lr := &mockLoyaltyRepo{tier: &domain.MembershipTier{Multiplier: 2.0}}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.EarnPoints(context.Background(), "", "c1", &txID, 50000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PointsDelta != 100 {
			t.Errorf("want 100 points, got %.2f", resp.PointsDelta)
		}
	})

	t.Run("no tier assigned defaults to 1x multiplier", func(t *testing.T) {
		lr := &mockLoyaltyRepo{tier: nil}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.EarnPoints(context.Background(), "", "c1", nil, 10000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PointsDelta != 10 {
			t.Errorf("want 10 points, got %.2f", resp.PointsDelta)
		}
	})

	t.Run("zero total returns zero-delta response without DB write", func(t *testing.T) {
		earnCalled := false
		lr := &mockLoyaltyRepoWithEarnTracker{
			mockLoyaltyRepo: mockLoyaltyRepo{
				tier: &domain.MembershipTier{Multiplier: 2.0},
			},
			onEarnCalled: func() { earnCalled = true },
		}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.EarnPoints(context.Background(), "", "c1", nil, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PointsDelta != 0 {
			t.Errorf("want 0 points, got %.2f", resp.PointsDelta)
		}
		if earnCalled {
			t.Error("EarnPoints DB call should not be made for zero points")
		}
	})

	t.Run("total below minimum unit earns zero points", func(t *testing.T) {
		lr := &mockLoyaltyRepo{tier: &domain.MembershipTier{Multiplier: 1.0}}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.EarnPoints(context.Background(), "", "c1", nil, 999)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PointsDelta != 0 {
			t.Errorf("want 0 points for total < 1000, got %.2f", resp.PointsDelta)
		}
	})

	t.Run("DB error on EarnPoints is propagated", func(t *testing.T) {
		lr := &mockLoyaltyRepo{
			tier:    &domain.MembershipTier{Multiplier: 1.0},
			earnErr: errors.New("db rollback"),
		}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.EarnPoints(context.Background(), "", "c1", &txID, 50000)
		if err == nil {
			t.Fatal("expected DB error to propagate")
		}
	})

	t.Run("tier fetch DB error halts earn", func(t *testing.T) {
		lr := &mockLoyaltyRepo{tierErr: errors.New("tier query failed")}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.EarnPoints(context.Background(), "", "c1", nil, 50000)
		if err == nil {
			t.Fatal("expected tier fetch error")
		}
	})
}

// ─── RedeemPoints ─────────────────────────────────────────────────────────────

//nolint:gocognit,funlen // test function with many subtests is inherently long
func TestLoyaltyService_RedeemPoints(t *testing.T) {
	nop := zerolog.Nop()
	txID := "txn-xyz"

	t.Run("successful redemption within balance", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 500}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.RedeemPoints(context.Background(), "c1", &txID, 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Type != "SPEND" {
			t.Error("expected SPEND type")
		}
		if resp.PointsDelta != -200 {
			t.Errorf("want -200 delta, got %.2f", resp.PointsDelta)
		}
	})

	t.Run("exact balance redemption allowed", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 100}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 100)
		if err != nil {
			t.Fatalf("exact balance redemption should succeed: %v", err)
		}
	})

	t.Run("redemption exceeding balance returns ErrInsufficientPoints", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 50}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 100)
		if err == nil {
			t.Fatal("expected ErrInsufficientPoints")
		}
		if !errors.Is(err, ErrInsufficientPoints) {
			t.Errorf("expected ErrInsufficientPoints, got: %v", err)
		}
	})

	t.Run("zero redemption returns ErrInvalidRedemption", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 500}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 0)
		if !errors.Is(err, ErrInvalidRedemption) {
			t.Errorf("expected ErrInvalidRedemption, got: %v", err)
		}
	})

	t.Run("negative redemption returns ErrInvalidRedemption", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 500}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, -50)
		if !errors.Is(err, ErrInvalidRedemption) {
			t.Errorf("expected ErrInvalidRedemption, got: %v", err)
		}
	})

	t.Run("zero balance with any positive redemption fails", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balance: 0}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 1)
		if !errors.Is(err, ErrInsufficientPoints) {
			t.Errorf("expected ErrInsufficientPoints, got: %v", err)
		}
	})

	t.Run("balance fetch DB error propagated", func(t *testing.T) {
		lr := &mockLoyaltyRepo{balanceErr: errors.New("db timeout")}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 100)
		if err == nil {
			t.Fatal("expected error from balance fetch")
		}
	})

	t.Run("DB error on SpendPoints (after balance check) is propagated — simulates rollback", func(t *testing.T) {
		lr := &mockLoyaltyRepo{
			balance:  500,
			spendErr: errors.New("transaction rolled back"),
		}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.RedeemPoints(context.Background(), "c1", nil, 100)
		if err == nil {
			t.Fatal("expected DB rollback error to propagate")
		}
	})
}

// ─── GetHistory ───────────────────────────────────────────────────────────────

func TestLoyaltyService_GetHistory(t *testing.T) {
	nop := zerolog.Nop()

	t.Run("returns entries newest first", func(t *testing.T) {
		now := time.Now()
		lr := &mockLoyaltyRepo{
			history: []*domain.LoyaltyLedger{
				{ID: "e2", CustomerID: "c1", PointsDelta: -50, Type: "SPEND", CreatedAt: now},
				{ID: "e1", CustomerID: "c1", PointsDelta: 100, Type: "EARN", CreatedAt: now.Add(-time.Hour)},
			},
		}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		resp, err := svc.GetHistory(context.Background(), "c1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("want 2 entries, got %d", len(resp))
		}
		if resp[0].Type != "SPEND" {
			t.Error("expected SPEND entry first (newest)")
		}
	})

	t.Run("history repo error propagated", func(t *testing.T) {
		lr := &mockLoyaltyRepo{historyErr: errors.New("db error")}
		svc := NewLoyaltyService(lr, &mockTierRepo{}, &mockCustomerRepoLoyalty{}, nil, nop)
		_, err := svc.GetHistory(context.Background(), "c1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ─── AssignTier ───────────────────────────────────────────────────────────────

func TestLoyaltyService_AssignTier(t *testing.T) {
	nop := zerolog.Nop()

	t.Run("assigns tier successfully", func(t *testing.T) {
		tierRepo := &mockTierRepo{findByID: &domain.MembershipTier{ID: "t1", Name: "Gold", Multiplier: 2.0}}
		lr := &mockLoyaltyRepo{}
		svc := NewLoyaltyService(lr, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		err := svc.AssignTier(context.Background(), "c1", "t1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unknown tier returns ErrTierNotFound", func(t *testing.T) {
		tierRepo := &mockTierRepo{findByID: nil}
		svc := NewLoyaltyService(&mockLoyaltyRepo{}, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		err := svc.AssignTier(context.Background(), "c1", "nonexistent")
		if !errors.Is(err, ErrTierNotFound) {
			t.Errorf("expected ErrTierNotFound, got: %v", err)
		}
	})

	t.Run("tier repo DB error propagated", func(t *testing.T) {
		tierRepo := &mockTierRepo{byIDErr: errors.New("db error")}
		svc := NewLoyaltyService(&mockLoyaltyRepo{}, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		err := svc.AssignTier(context.Background(), "c1", "t1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("loyalty repo assign error propagated", func(t *testing.T) {
		tierRepo := &mockTierRepo{findByID: &domain.MembershipTier{ID: "t1", Name: "Bronze", Multiplier: 1.0}}
		lr := &mockLoyaltyRepo{assignErr: errors.New("customer not found")}
		svc := NewLoyaltyService(lr, tierRepo, &mockCustomerRepoLoyalty{}, nil, nop)
		err := svc.AssignTier(context.Background(), "c999", "t1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
