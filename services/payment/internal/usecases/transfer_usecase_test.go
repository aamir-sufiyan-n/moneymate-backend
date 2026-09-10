package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authclient "github.com/moneymate-2026/moneymate-backend/services/payment/internal/adapter/authClient"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/usecases"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type mockAccountRepo struct {
	domain.AccountRepository
	accounts map[uuid.UUID]*domain.Account
	wallets  map[uuid.UUID]*domain.Account
}

func (m *mockAccountRepo) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Account, error) {
	if acc, ok := m.wallets[userID]; ok {
		return acc, nil
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	if acc, ok := m.accounts[id]; ok {
		return acc, nil
	}
	return nil, apperrors.ErrNotFound
}

type mockTxRepo struct {
	domain.TransactionRepository
	txs                 []*domain.Transaction
	totalCount          int64
	err                 error
	capturedAccountID   uuid.UUID
	capturedCategoryID *uuid.UUID
	capturedLimit       int32
	capturedOffset      int32
}

func (m *mockTxRepo) ListByAccountPaginated(ctx context.Context, accountID uuid.UUID, categoryID *uuid.UUID, limit, offset int32) ([]*domain.Transaction, int64, error) {
	m.capturedAccountID = accountID
	m.capturedCategoryID = categoryID
	m.capturedLimit = limit
	m.capturedOffset = offset

	if m.err != nil {
		return nil, 0, m.err
	}
	return m.txs, m.totalCount, nil
}

type mockCategoryRepo struct {
	domain.CategoryRepository
	categories map[uuid.UUID]*domain.Category
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	if cat, ok := m.categories[id]; ok {
		return cat, nil
	}
	return nil, apperrors.ErrNotFound
}

type mockAuthClient struct {
	profiles map[string]*authclient.UserProfile
}

func (m *mockAuthClient) GetUserProfile(ctx context.Context, userID string) (*authclient.UserProfile, error) {
	if p, ok := m.profiles[userID]; ok {
		return p, nil
	}
	return nil, apperrors.ErrNotFound
}

type mockMerchantClient struct{}

func (m *mockMerchantClient) GetStoreProfile(ctx context.Context, storeID string) (string, string, error) {
	return "", "", nil
}

func TestListMyTransactions_WithCategoryFilter(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	catID := uuid.New()
	catIDStr := catID.String()

	acc := &domain.Account{
		ID:     walletID,
		UserID: &userID,
		Type:   domain.AccountTypeWallet,
	}

	accRepo := &mockAccountRepo{
		accounts: map[uuid.UUID]*domain.Account{walletID: acc},
		wallets:  map[uuid.UUID]*domain.Account{userID: acc},
	}

	now := time.Now().UTC()
	tx1 := &domain.Transaction{
		ID:            uuid.New(),
		FromAccountID: walletID,
		ToAccountID:   uuid.New(),
		Amount:        5000,
		Status:        domain.TxStatusCompleted,
		CategoryID:    &catID,
		CreatedAt:     now,
	}

	txRepo := &mockTxRepo{
		txs:        []*domain.Transaction{tx1},
		totalCount: 1,
	}

	catRepo := &mockCategoryRepo{
		categories: map[uuid.UUID]*domain.Category{
			catID: {ID: catID, UserID: userID, Name: "Groceries"},
		},
	}

	authClient := &mockAuthClient{
		profiles: map[string]*authclient.UserProfile{
			userID.String(): {FullName: "Alice", Handle: "alice"},
		},
	}

	uc := usecases.NewTransferUsecase(accRepo, txRepo, nil, catRepo, authClient, &mockMerchantClient{})

	t.Run("success with category filter", func(t *testing.T) {
		res, err := uc.ListMyTransactions(ctx, usecases.ListTransactionsInput{
			AuthenticatedUserID: userID.String(),
			CategoryID:          &catIDStr,
			Page:                1,
			PageSize:            10,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if txRepo.capturedAccountID != walletID {
			t.Errorf("expected account ID %v, got %v", walletID, txRepo.capturedAccountID)
		}
		if txRepo.capturedCategoryID == nil || *txRepo.capturedCategoryID != catID {
			t.Errorf("expected category ID %v, got %v", catID, txRepo.capturedCategoryID)
		}
		if txRepo.capturedLimit != 10 || txRepo.capturedOffset != 0 {
			t.Errorf("expected limit 10, offset 0, got limit %d, offset %d", txRepo.capturedLimit, txRepo.capturedOffset)
		}
		if res.TotalCount != 1 || len(res.Transactions) != 1 {
			t.Fatalf("expected 1 transaction, got total_count %d, len %d", res.TotalCount, len(res.Transactions))
		}
		if res.Transactions[0].Category != "Groceries" {
			t.Errorf("expected category Groceries, got %s", res.Transactions[0].Category)
		}
	})

	t.Run("success without category filter", func(t *testing.T) {
		_, err := uc.ListMyTransactions(ctx, usecases.ListTransactionsInput{
			AuthenticatedUserID: userID.String(),
			CategoryID:          nil,
			Page:                1,
			PageSize:            10,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if txRepo.capturedCategoryID != nil {
			t.Errorf("expected nil category ID, got %v", txRepo.capturedCategoryID)
		}
	})

	t.Run("invalid category uuid returns error", func(t *testing.T) {
		invalidCat := "invalid-uuid"
		_, err := uc.ListMyTransactions(ctx, usecases.ListTransactionsInput{
			AuthenticatedUserID: userID.String(),
			CategoryID:          &invalidCat,
			Page:                1,
			PageSize:            10,
		})
		if err != apperrors.ErrInvalidInput {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})
}
