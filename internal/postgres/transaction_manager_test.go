package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestNewTransactionManager(t *testing.T) {
	t.Parallel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	if tm == nil {
		t.Fatal("NewTransactionManager returned nil")
	}
}

func TestTxFromContext_returnsNilWhenOutsideTransaction(t *testing.T) {
	t.Parallel()
	tx, ok := TxFromContext(context.Background())
	if tx != nil || ok {
		t.Errorf("TxFromContext() = %v, %v; want nil, false", tx, ok)
	}
}

func TestTxFromContext_returnsTxWhenInsideTransaction(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	var capturedTx any

	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		tx, ok := TxFromContext(ctx)
		if !ok || tx == nil {
			t.Errorf("TxFromContext inside transaction: tx=%v, ok=%v", tx, ok)
			return errors.New("no tx in context")
		}
		capturedTx = tx
		return nil
	})

	if err != nil {
		t.Fatalf("WithinRWTransaction: %v", err)
	}
	if capturedTx == nil {
		t.Error("tx was not captured from context")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinROTransaction_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	called := false
	err = tm.WithinROTransaction(context.Background(), func(ctx context.Context) error {
		called = true
		tx, ok := TxFromContext(ctx)
		if !ok || tx == nil {
			t.Error("expected tx in context")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WithinROTransaction: %v", err)
	}
	if !called {
		t.Error("callback was not called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinRWTransaction_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	called := false
	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("WithinRWTransaction: %v", err)
	}
	if !called {
		t.Error("callback was not called")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinRWTransaction_nestedUsesExistingTx(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Only one Begin and one Commit: inner call must not start a new transaction
	mock.ExpectBegin()
	mock.ExpectCommit()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	innerCalled := false
	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		return tm.WithinRWTransaction(ctx, func(innerCtx context.Context) error {
			innerCalled = true
			tx, ok := TxFromContext(innerCtx)
			if !ok || tx == nil {
				t.Error("nested callback should see same tx in context")
			}
			return nil
		})
	})

	if err != nil {
		t.Fatalf("nested WithinRWTransaction: %v", err)
	}
	if !innerCalled {
		t.Error("inner callback was not called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinRWTransaction_rollbackOnCallbackError(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	wantErr := errors.New("callback failed")
	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("WithinRWTransaction err = %v; want %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinRWTransaction_rollbackAndRethrowPanic(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to be rethrown")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	}()

	_ = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		panic("test panic")
	})
}

func TestWithinRWTransaction_beginFails(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	beginErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(beginErr)

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	called := false
	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err == nil {
		t.Fatal("expected error from BeginTxx")
	}
	if !errors.Is(err, beginErr) {
		t.Errorf("expected wrapped begin error (errors.Is), got: %v", err)
	}
	if err.Error() == "" || len(err.Error()) < 10 || err.Error()[:10] != "begin tx: " {
		t.Errorf("expected 'begin tx: ...' message, got: %v", err)
	}
	if called {
		t.Error("callback must not be called when Begin fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestWithinRWTransaction_commitFails(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	commitErr := errors.New("commit failed")
	mock.ExpectCommit().WillReturnError(commitErr)

	sqlxDB := sqlx.NewDb(db, "pgx")
	tm := NewTransactionManager(sqlxDB)

	err = tm.WithinRWTransaction(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err == nil {
		t.Fatal("expected error from Commit")
	}
	if err.Error() == "" {
		t.Errorf("expected commit error message, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
