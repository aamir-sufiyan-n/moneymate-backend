package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/payment/sqlc/generated"
)

type TransactionRepo struct {
	q *generated.Queries
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{q: generated.New(pool)}
}

func (r *TransactionRepo) Create(ctx context.Context, t *domain.Transaction) error {
	_, err := r.q.InsertTransaction(ctx, generated.InsertTransactionParams{
		ID:             t.ID,
		FromAccountID:  t.FromAccountID,
		ToAccountID:    t.ToAccountID,
		Amount:         t.Amount,
		Column5:        generated.PaymentTxStatus(t.Status),
		IdempotencyKey: t.IdempotencyKey,
		Description:    &t.Description,
		CompletedAt:    timePtrToPgtype(t.CompletedAt),
	})
	return mapDBErr(err)
}

func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, mapDBErr(err)
	}
	return rowToTransaction(row), nil
}

func (r *TransactionRepo) GetByIdempotencyKey(ctx context.Context, key string, fromAccountID uuid.UUID) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByIdempotencyKey(ctx, generated.GetTransactionByIdempotencyKeyParams{
		IdempotencyKey: key,
		FromAccountID:  fromAccountID,
	})
	if err != nil {
		return nil, mapDBErr(err)
	}
	return rowToTransaction(generated.GetTransactionByIDRow(row)), nil
}


func (r *TransactionRepo) ListByAccountPaginated(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]*domain.Transaction, int64, error) {
	rows, err := r.q.ListTransactionsByAccountPaginated(ctx, generated.ListTransactionsByAccountPaginatedParams{
		FromAccountID: accountID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, mapDBErr(err)
	}
	total, err := r.q.CountTransactionsByAccount(ctx, accountID)
	if err != nil {
		return nil, 0, mapDBErr(err)
	}
	txs := make([]*domain.Transaction, len(rows))
	for i, row := range rows {
		txs[i] = paymentTxToTransaction(row)
	}
	return txs, total, nil
}

func (r *TransactionRepo) ListWithdrawals(ctx context.Context, settlementAccountID uuid.UUID, fromAccountID *uuid.UUID, limit, offset int32) ([]*domain.Transaction, int64, error) {
	var fromParam pgtype.UUID
	if fromAccountID != nil {
		fromParam = pgtype.UUID{Bytes: *fromAccountID, Valid: true}
	}
	rows, err := r.q.ListWithdrawals(ctx, generated.ListWithdrawalsParams{
		Limit: limit, Offset: offset,
		SettlementAccountID: settlementAccountID, FromAccountID: fromParam,
	})
	if err != nil {
		return nil, 0, mapDBErr(err)
	}
	total, err := r.q.CountWithdrawals(ctx, generated.CountWithdrawalsParams{
		SettlementAccountID: settlementAccountID, FromAccountID: fromParam,
	})
	if err != nil {
		return nil, 0, mapDBErr(err)
	}
	txs := make([]*domain.Transaction, len(rows))
	for i, row := range rows {
		txs[i] = paymentTxToTransaction(row)
	}
	return txs, total, nil
}


func (r *TransactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TxStatus) error {
	return mapDBErr(r.q.UpdateTransactionStatus(ctx, generated.UpdateTransactionStatusParams{
		ID:      id,
		Column2: generated.PaymentTxStatus(status),
	}))
}

func (r *TransactionRepo) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactionsByAccount(ctx, accountID)
	if err != nil {
		return nil, mapDBErr(err)
	}
	txs := make([]*domain.Transaction, 0, len(rows))
	for _, row := range rows {
		txs = append(txs, rowToTransaction(generated.GetTransactionByIDRow(row)))
	}
	return txs, nil
}

func (r *TransactionRepo) GetEntriesByTransactionID(ctx context.Context, txID uuid.UUID) ([]*domain.JournalEntry, error) {
	rows, err := r.q.GetEntriesByTransactionID(ctx, txID)
	if err != nil {
		return nil, mapDBErr(err)
	}
	entries := make([]*domain.JournalEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &domain.JournalEntry{
			ID:            row.ID,
			TransactionID: row.TransactionID,
			AccountID:     row.AccountID,
			Amount:        row.Amount,
			Direction:     row.Direction,
			CreatedAt:     row.CreatedAt,
		})
	}
	return entries, nil
}

func rowToTransaction(row generated.GetTransactionByIDRow) *domain.Transaction {
	return &domain.Transaction{
		ID:             row.ID,
		FromAccountID:  row.FromAccountID,
		ToAccountID:    row.ToAccountID,
		Amount:         row.Amount,
		Status:         domain.TxStatus(row.Status),
		IdempotencyKey: row.IdempotencyKey,
		Description:    row.Description,
		CreatedAt:      row.CreatedAt,
		CompletedAt:    pgtypeToTimePtr(row.CompletedAt),
	}
}


func paymentTxToTransaction(row generated.PaymentTransaction) *domain.Transaction {
	var desc string
	if row.Description != nil {
		desc = *row.Description
	}
	return &domain.Transaction{
		ID:             row.ID,
		FromAccountID:  row.FromAccountID,
		ToAccountID:    row.ToAccountID,
		Amount:         row.Amount,
		Status:         domain.TxStatus(row.Status), // PaymentTxStatus -> domain.TxStatus, both string-backed enums, safe cast
		IdempotencyKey: row.IdempotencyKey,
		Description:    desc,
		CreatedAt:      row.CreatedAt,
		CompletedAt:    pgtypeToTimePtr(row.CompletedAt),
	}
}