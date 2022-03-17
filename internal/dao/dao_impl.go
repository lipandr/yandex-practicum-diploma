package dao

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lipandr/yandex-practicum-diploma/internal/types"
)

// NewUser метод DAO добавления нового пользователя.
func (d *DAO) NewUser(userID, encPass string) (int, error) {
	var id int
	err := d.dao.QueryRow(
		"INSERT INTO users (login, encrypted_password) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id;",
		userID, encPass).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUserByLogin метод DAO получения записи о пользователе.
func (d *DAO) GetUserByLogin(login string) (*types.TUser, error) {
	var u types.TUser
	err := d.dao.QueryRow(
		"SELECT id, encrypted_password FROM users WHERE login = ($1)", login).Scan(&u.ID, &u.EncryptedPassword)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SaveToken метод DAO сохранения токена выданного пользователю.
func (d *DAO) SaveToken(userID int, token string) error {
	_, err := d.dao.Exec(
		"INSERT INTO tokens (user_id, token) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET token = ($2)",
		userID, token)
	if err != nil {
		return err
	}
	return nil
}

// GetToken метод DAO получения сохраненного токена пользователя.
func (d *DAO) GetToken(token string) (int, error) {
	var userID int
	err := d.dao.QueryRow(
		"SELECT user_id FROM tokens WHERE token = ($1)", token).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// NewOrder метод DAO сохранения нового заказа для расчета начислений.
func (d *DAO) NewOrder(userID int, orderNumber string) error {
	_, err := d.dao.Exec(
		"INSERT INTO orders (order_number, user_id, status) "+
			"VALUES ($1, $2, $3);",
		orderNumber, userID, "NEW")
	if err != nil {
		return err
	}
	return nil
}

// IsOrderExists метод DAO проверки сохраненного заказа.
func (d *DAO) IsOrderExists(userID int, orderNumber string) error {
	var o types.Order
	err := d.dao.QueryRow(
		"SELECT order_number, user_id FROM orders WHERE order_number = ($1)", orderNumber).
		Scan(&o.OrderNumber, &o.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if o.OrderNumber == orderNumber {
		if o.UserID == userID {
			return types.ErrOrderUploadedByUser
		} else {
			return types.ErrOrderUploadedByOtherUser
		}
	}
	return nil
}

// IsOrderWithdrawn метод DAO проверки осуществленных списаний по номеру заказа.
func (d *DAO) IsOrderWithdrawn(orderNumber string) error {
	var o string
	err := d.dao.QueryRow(
		"SELECT order_number FROM withdraws WHERE order_number = ($1)", orderNumber).
		Scan(&o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if o == orderNumber {
		return types.ErrOrderAlreadyWithdrawn
	}
	return nil
}

// GetOrderList метод DAO получения списка заказов пользователя.
func (d *DAO) GetOrderList(userID int) ([]types.Order, error) {
	var orders []types.Order
	rows, err := d.dao.Query(
		"SELECT order_number, status, accrual, uploaded_at from orders where user_id = ($1) ORDER BY uploaded_at ;", userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var o types.Order
		var t time.Time
		err = rows.Scan(&o.OrderNumber, &o.Status, &o.Accrual, &t)
		if err != nil {
			return nil, err
		}
		o.UploadedAt = t.Local().Format(time.RFC3339)
		orders = append(orders, o)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetTotalWithdrawals метод DAO получения суммы списаний, осуществленных пользователем.
func (d *DAO) GetTotalWithdrawals(userID int) (float64, error) {
	var w float64
	err := d.dao.QueryRow(
		"SELECT coalesce(SUM(sum), 0.00) FROM withdraws WHERE user_id = ($1)", userID).
		Scan(&w)
	if err != nil {
		return 0, err
	}
	return w, nil
}

// GetAccruals метод DAO получения суммы начислений пользователя.
func (d *DAO) GetAccruals(userID int) (float64, error) {
	var a float64
	err := d.dao.QueryRow(
		"SELECT SUM(accrual) FROM orders WHERE user_id = ($1)", userID).
		Scan(&a)
	if err != nil {
		return 0, err
	}
	return a, nil
}

// NewWithdrawal метод DAO добавления нового списания пользователя.
func (d *DAO) NewWithdrawal(userID int, sum float64, orderNumber string) error {
	_, err := d.dao.Exec(
		"INSERT INTO withdraws (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4);",
		userID, orderNumber, sum, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// GetWithdrawalsList метод DAO получения списка списаний пользователя.
func (d *DAO) GetWithdrawalsList(userID int) ([]types.Withdraw, error) {
	var wthd []types.Withdraw
	rows, err := d.dao.Query(
		"SELECT order_number, sum, processed_at from withdraws where user_id = ($1) ORDER BY processed_at ;", userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var w types.Withdraw
		var t time.Time
		err = rows.Scan(&w.OrderNumber, &w.Sum, &t)
		if err != nil {
			return nil, err
		}
		w.ProcessedAt = t.Format(time.RFC3339)
		wthd = append(wthd, w)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return wthd, nil
}

// GetOrdersForProcessing метод DAO получения списка заказов для расчета начислений.
func (d *DAO) GetOrdersForProcessing(wps int) ([]string, error) {
	var orders []string
	rows, err := d.dao.Query(
		"SELECT order_number FROM orders WHERE status IN ($1, $2) ORDER BY uploaded_at LIMIT $3", "NEW", "PROCESSING", wps,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			return orders, err
		}
		orders = append(orders, orderID)
	}
	err = rows.Err()
	return orders, err
}

// UpdateOrderState метод DAO обновления статуса заказа по результатам расчета начислений.
func (d *DAO) UpdateOrderState(status *types.AccrualOrderState) error {
	_, err := d.dao.Exec(
		"UPDATE orders SET status=$1, accrual=$2 WHERE order_number = ($3)",
		status.Status, status.Accrual, status.Order,
	)
	return err
}
