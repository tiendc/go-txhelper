[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# Golang SQL Tx helper library

Execute SQL transaction with convenience.

## Installation

```shell
go get github.com/tiendc/go-txhelper
```

## Usage

```go
err := txhelper.Execute(ctx, db, func (db *sql.Tx) error {
    // Perform SQL statements with the inner `db`
})
if err != nil {
    fmt.Println(err)
}
```

### Set isolation level

```go
err := txhelper.Execute(ctx, db, func (db *sql.Tx) error {
    // Perform SQL statements with the inner `db`
}, txhelper.IsolationLevel(sql.LevelReadCommitted))
```

### Retry on deadlock

Normally a deadlock error can be retried instead of returning error.

```go
// For MySQL and when you use driver "github.com/go-sql-driver/mysql"

import "github.com/go-sql-driver/mysql"
func IsMySQLDeadlock(err error) bool {
    sqlErr := &mysql.MySQLError{}
    // Error 1213: Deadlock found when trying to get lock
    // Error 1205: Lock wait timeout exceeded
    return errors.As(err, &sqlErr) && (sqlErr.Number == 1213 || sqlErr.Number == 1205)
}

err := txhelper.Execute(ctx, db, func (db *sql.Tx) error {
    // Perform SQL statements with the inner `db`
}, txhelper.CheckRetryable(IsMySQLDeadlock),
    txhelper.MaxRetryTimes(3),            // optional, default is 3
    txhelper.RetryDelay(3 * time.Second), // optional, default is 0 - no delay
)
```

```go
// For Postgres and when you use driver "github.com/lib/pq"

import "github.com/lib/pq"
func IsPostgresDeadlock(err error) bool {
    sqlErr := &pq.Error{}
    return errors.As(err, &sqlErr) && (sqlErr.Code == "40P01")
}

err := txhelper.Execute(ctx, db, func (db *sql.Tx) error {
    // Perform SQL statements with the inner `db`
}, txhelper.CheckRetryable(IsPostgresDeadlock))
```

## Contributing

- You are welcome to make pull requests for new functions and bug fixes.

## License

- [MIT License](LICENSE)

[doc-img]: https://pkg.go.dev/badge/github.com/tiendc/go-txhelper
[doc]: https://pkg.go.dev/github.com/tiendc/go-txhelper
[gover-img]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[gover]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[ci-img]: https://github.com/tiendc/go-txhelper/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/tiendc/go-txhelper/actions/workflows/go.yml
[cov-img]: https://codecov.io/gh/tiendc/go-txhelper/branch/main/graph/badge.svg
[cov]: https://codecov.io/gh/tiendc/go-txhelper
[rpt-img]: https://goreportcard.com/badge/github.com/tiendc/go-txhelper
[rpt]: https://goreportcard.com/report/github.com/tiendc/go-txhelper
