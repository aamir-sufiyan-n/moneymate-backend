package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type mockAccountRepo struct {
	domain.AccountRepository
	wallet *domain.Account
	err    error
}

func (m *mockAccountRepo) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.wallet, nil
}

type mockTxRepo struct {
	domain.TransactionRepository
	categoryRows []domain.SpendByCategory
	periodRows   []domain.SpendByPeriod
	categoryErr  error
	periodErr    error

	capturedAccountID   uuid.UUID
	capturedFrom        time.Time
	capturedTo          time.Time
	capturedGranularity string
}

func (m *mockTxRepo) GetSpendByCategory(ctx context.Context, accountID uuid.UUID, from, to time.Time) ([]domain.SpendByCategory, error) {
	m.capturedAccountID = accountID
	m.capturedFrom = from
	m.capturedTo = to
	if m.categoryErr != nil {
		return nil, m.categoryErr
	}
	return m.categoryRows, nil
}

func (m *mockTxRepo) GetSpendByPeriod(ctx context.Context, accountID uuid.UUID, from, to time.Time, granularity string) ([]domain.SpendByPeriod, error) {
	m.capturedAccountID = accountID
	m.capturedFrom = from
	m.capturedTo = to
	m.capturedGranularity = granularity
	if m.periodErr != nil {
		return nil, m.periodErr
	}
	return m.periodRows, nil
}

func TestParseDateRange(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	t.Run("defaults when empty", func(t *testing.T) {
		from, to, err := parseDateRange("", "", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedFrom := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		expectedTo := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		if !from.Equal(expectedFrom) {
			t.Errorf("expected from %v, got %v", expectedFrom, from)
		}
		if !to.Equal(expectedTo) {
			t.Errorf("expected to %v, got %v", expectedTo, to)
		}
	})

	t.Run("precise to date exclusive upper boundary", func(t *testing.T) {
		from, to, err := parseDateRange("2026-09-01", "2026-09-09", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedFrom := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		expectedTo := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC) // full day covered!
		if !from.Equal(expectedFrom) {
			t.Errorf("expected from %v, got %v", expectedFrom, from)
		}
		if !to.Equal(expectedTo) {
			t.Errorf("expected to %v, got %v", expectedTo, to)
		}
	})

	t.Run("invalid to before from", func(t *testing.T) {
		_, _, err := parseDateRange("2026-09-10", "2026-09-01", now)
		if err == nil {
			t.Fatal("expected error when to <= from, got nil")
		}
	})
}

func TestSpendByCategory_Success(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	accountRepo := &mockAccountRepo{
		wallet: &domain.Account{
			ID:     walletID,
			UserID: &userID,
			Type:   domain.AccountTypeWallet,
		},
	}
	txRepo := &mockTxRepo{
		categoryRows: []domain.SpendByCategory{
			{Category: "food", TransactionCount: 5, TotalAmount: 125000},
			{Category: "hotel", TransactionCount: 2, TotalAmount: 40000},
		},
	}

	uc := NewAnalyticsUsecase(accountRepo, txRepo)
	res, err := uc.SpendByCategory(context.Background(), userID.String(), "2026-09-01", "2026-09-09")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if txRepo.capturedAccountID != walletID {
		t.Errorf("expected account ID %v, got %v", walletID, txRepo.capturedAccountID)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res))
	}
	if res[0].Category != "food" || res[0].TotalAmount != "1250.00" || res[0].TransactionCount != 5 {
		t.Errorf("unexpected res[0]: %+v", res[0])
	}
	if res[1].Category != "hotel" || res[1].TotalAmount != "400.00" || res[1].TransactionCount != 2 {
		t.Errorf("unexpected res[1]: %+v", res[1])
	}
}

func TestSpendByPeriod_ZeroFilledBuckets(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	accountRepo := &mockAccountRepo{
		wallet: &domain.Account{
			ID:     walletID,
			UserID: &userID,
			Type:   domain.AccountTypeWallet,
		},
	}
	txRepo := &mockTxRepo{
		periodRows: []domain.SpendByPeriod{
			{Period: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), TotalAmount: 85000, TransactionCount: 3},
		},
	}

	uc := NewAnalyticsUsecase(accountRepo, txRepo)
	res, err := uc.SpendByPeriod(context.Background(), userID.String(), "day", "2026-09-01", "2026-09-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Range 2026-09-01 to 2026-09-04 (03 inclusive) has 3 days: 2026-09-01, 2026-09-02, 2026-09-03
	if len(res) != 3 {
		t.Fatalf("expected 3 day buckets, got %d", len(res))
	}

	if res[0].Period != "2026-09-01" || res[0].TotalAmount != "0.00" || res[0].TransactionCount != 0 {
		t.Errorf("expected zero bucket for 2026-09-01, got %+v", res[0])
	}
	if res[1].Period != "2026-09-02" || res[1].TotalAmount != "850.00" || res[1].TransactionCount != 3 {
		t.Errorf("expected active bucket for 2026-09-02, got %+v", res[1])
	}
	if res[2].Period != "2026-09-03" || res[2].TotalAmount != "0.00" || res[2].TransactionCount != 0 {
		t.Errorf("expected zero bucket for 2026-09-03, got %+v", res[2])
	}
}

func TestSpendByPeriod_InvalidGranularity(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	accountRepo := &mockAccountRepo{
		wallet: &domain.Account{
			ID:     walletID,
			UserID: &userID,
			Type:   domain.AccountTypeWallet,
		},
	}
	txRepo := &mockTxRepo{}
	uc := NewAnalyticsUsecase(accountRepo, txRepo)

	_, err := uc.SpendByPeriod(context.Background(), userID.String(), "hourly", "2026-09-01", "2026-09-03")
	if err != apperrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
