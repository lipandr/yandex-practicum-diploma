package dao

const (
	// UsersTable таблица для хранения зарегистрированных пользователей.
	UsersTable = `
CREATE TABLE IF NOT EXISTS users
(
    id serial PRIMARY KEY,
    login text NOT NULL UNIQUE, 
	encrypted_password text
);
`
	// UserTokens таблица для хранения токенов авторизации выданных пользователям.
	UserTokens = `
CREATE TABLE IF NOT EXISTS tokens
(
	id serial PRIMARY KEY,
	user_id serial REFERENCES users(id) UNIQUE,
	token text,
	created_at timestamp without time zone default now()
);
`
	// OrdersTable таблица хранения номеров заказов для расчета начислений.
	OrdersTable = `
CREATE TABLE IF NOT EXISTS orders
(
    id  serial PRIMARY KEY,
    order_number text NOT NULL UNIQUE,
	user_id serial REFERENCES users(id),
	status text,
	accrual real,
	uploaded_at timestamp without time zone default now()
);
`
	// WithdrawsTable таблица хранения списаний пользователей.
	WithdrawsTable = `
CREATE TABLE IF NOT EXISTS withdraws
(
    id  serial PRIMARY KEY,
    user_id serial REFERENCES users(id),
	order_number text NOT NULL UNIQUE,
	sum real,
	processed_at timestamp without time zone default now()
);
`
)
