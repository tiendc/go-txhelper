package txhelper

import (
	"context"
	"database/sql"
	"time"
)

type (
	TxBeginner interface {
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	}

	TxExecFunc         func(db *sql.Tx) error
	CheckRetryableFunc func(error) bool

	options struct {
		isolation      sql.IsolationLevel
		retryTimes     uint
		maxRetryTimes  uint
		retryDelay     time.Duration
		checkRetryable CheckRetryableFunc
	}
)

var (
	defaultCheckRetryable CheckRetryableFunc
)

func defaultOptions() *options {
	return &options{
		isolation:      sql.LevelDefault,
		retryTimes:     0,
		maxRetryTimes:  3, //nolint:mnd
		retryDelay:     0,
		checkRetryable: defaultCheckRetryable,
	}
}

func MaxRetryTimes(times uint) func(*options) {
	return func(r *options) {
		r.maxRetryTimes = times
	}
}

func RetryDelay(delay time.Duration) func(*options) {
	return func(r *options) {
		r.retryDelay = delay
	}
}

func IsolationLevel(level sql.IsolationLevel) func(*options) {
	return func(r *options) {
		r.isolation = level
	}
}

func CheckRetryable(fn CheckRetryableFunc) func(*options) {
	return func(r *options) {
		r.checkRetryable = fn
	}
}

func DefaultCheckRetryable() CheckRetryableFunc {
	return defaultCheckRetryable
}

func SetDefaultCheckRetryable(fn CheckRetryableFunc) {
	defaultCheckRetryable = fn
}

// Execute executes function within a transaction
func Execute(ctx context.Context, db TxBeginner, exec TxExecFunc, opts ...func(*options)) error {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: cfg.isolation,
	})
	if err != nil {
		return err
	}

	// Make sure we won't miss panicking from exec function
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	for {
		err = exec(tx)
		if err == nil {
			err = tx.Commit()
		}

		if err != nil && cfg.checkRetryable != nil && cfg.retryTimes < cfg.maxRetryTimes && cfg.checkRetryable(err) {
			if cfg.retryDelay > 0 {
				time.Sleep(cfg.retryDelay)
			}
			cfg.retryTimes++
			continue
		}

		// Rollback if error or commit failed
		if err != nil {
			// Don't reassign err here as if rollback succeeds, no error returns
			_ = tx.Rollback()
		}

		return err
	}
}
