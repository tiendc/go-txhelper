package txhelper

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	errRetryable = errors.New("deadlock error")
	errSQL       = errors.New("sql error")
)

func isRetryableErr(err error) bool {
	return errors.Is(err, errRetryable)
}

func init() {
	SetDefaultCheckRetryable(isRetryableErr)
}

func Test_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expect no error, got %v", err)
		}
	})

	t.Run("Failure: error on Begin", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin().WillReturnError(errSQL)
		mockDB.ExpectRollback()

		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			return nil
		})
		if err == nil || !errors.Is(err, errSQL) {
			t.Errorf("Expect has error %v, got %v", errSQL, err)
		}
	})

	t.Run("Failure: error on execution func", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			return errSQL
		})
		if err == nil || !errors.Is(err, errSQL) {
			t.Errorf("Expect has error %v, got %v", errSQL, err)
		}
	})

	t.Run("Failure: panic on execution func", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		defer func() {
			err := recover()
			if err == nil || !errors.Is(err.(error), errSQL) {
				t.Errorf("Expect has error %v, got %v", errSQL, err)
			}
		}()

		_ = Execute(context.Background(), db, func(tx *sql.Tx) error {
			panic(errSQL)
		})
	})
}

func Test_Execute_WithOptions(t *testing.T) {
	t.Run("Success with retry 3 times", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		count := 0
		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			if count == 3 {
				return nil
			}
			count++
			return errRetryable
		}, RetryDelay(200*time.Millisecond))
		if err != nil {
			t.Errorf("Expect no error, got %v", err)
		}

		if count != 3 {
			t.Errorf("Expect retry 3 times, got %d", count)
		}
	})

	t.Run("Failure with maximum 7 retry times", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		count := 0
		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			count++
			return errRetryable
		}, MaxRetryTimes(7))
		if err == nil {
			t.Errorf("Expect got error, got nil")
		}

		if count != 8 {
			t.Errorf("Expect retry 7 times, got %d", count)
		}
	})

	t.Run("Failure non-retryable error", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		count := 0
		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			count++
			return errSQL
		})
		if err == nil {
			t.Errorf("Expect got error, got nil")
		}

		if count != 1 {
			t.Errorf("Expect no retry, got %d", count)
		}
	})

	t.Run("Success set Isolation LevelReadCommitted", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		err := Execute(context.Background(), db, func(tx *sql.Tx) error {
			return nil
		}, IsolationLevel(sql.LevelReadCommitted))
		if err != nil {
			t.Errorf("Expect no error, got %v", err)
		}
	})
}
