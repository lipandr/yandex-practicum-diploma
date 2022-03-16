package dao

import (
	"database/sql"
	_ "github.com/jackc/pgx"
	_ "github.com/lib/pq"
)

type DAO struct {
	dao *sql.DB
}

// NewDAO открытие соединения с БД и создание таблиц.
func NewDAO(dataSourceName string) (*DAO, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	tables := []string{
		UsersTable,
		OrdersTable,
		WithdrawsTable,
		UserTokens,
	}
	for _, table := range tables {
		if _, err = db.Exec(table); err != nil {
			return nil, err
		}
	}
	return &DAO{
		dao: db,
	}, nil
}
