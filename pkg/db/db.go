package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Client клиент для работы с бд
type Client interface {
	DB() DB
	Close() error
}

// DB интерфейс для работы с бд
type DB interface {
	SQLExecer
	Pinger
	Transactor
	Close()
}

// Query обертка над запросом, хранящая имя запроса и сам запрос
// имя запроса используется для логирования и потенциально может использоваться еще где-то, например в трейсинге.
type Query struct {
	Name     string
	QueryRow string
}

// Transactor интерфейс для работы с транзакциями.
type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// TxManager менеджер транзакций.
type TxManager interface {
	ReadCommitted(ctx context.Context, f Handler) error
}

// Handler функция, которая выполняется в транзакции.
type Handler func(ctx context.Context) error

// SQLExecer комбинирует NamedExecer и QueryExecer
type SQLExecer interface {
	NamedExecer
	QueryExecer
}

// NamedExecer интерфейс для работы с именованными запросами с помощью тегов в структурах.
type NamedExecer interface {
	ScanOneContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
	ScanAllContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
}

// QueryExecer интерфейс для работы с обычными запросами.
type QueryExecer interface {
	ExecContext(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row
}

// Pinger интерфейс для проверки соединения с бд
type Pinger interface {
	Ping(ctx context.Context) error
}
