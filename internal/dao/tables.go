package dao

const (
	UsersTable = `
CREATE TABLE IF NOT EXISTS users
(
    id serial PRIMARY KEY,
    login text NOT NULL UNIQUE, 
	encrypted_password text
);
`
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
	WithdrawsTable = `
CREATE TABLE IF NOT EXISTS withdraws
(
    id  serial PRIMARY KEY,
    user_id serial REFERENCES users(id),
	order_number text NOT NULL UNIQUE,
	sum int,
	processed_at timestamp without time zone default now()
);
`
	UserTokens = `
CREATE TABLE IF NOT EXISTS tokens
(
	id serial PRIMARY KEY,
	user_id serial REFERENCES users(id) UNIQUE,
	token text,
	created_at timestamp without time zone default now()
);
`
)
