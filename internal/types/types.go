package types

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
)

const (
	WorkersPoolSize int         = 10
	UserID          UserSession = "userID"
)

var (
	ErrUsersNotAuthenticated    = errors.New("user is not authenticated")
	ErrUsersAlreadyExists       = errors.New("user already exists")
	ErrOrderUploadedByUser      = errors.New("order uploaded user")
	ErrOrderUploadedByOtherUser = errors.New("order uploaded by other user")
	ErrOrderAlreadyWithdrawn    = errors.New("order already withdrawn")
	ErrInsufficientAccruals     = errors.New("insufficient accruals on the account")
	ErrOrderNumberInvalid       = errors.New("invalid order number")
)

type UserSession string

type TUser struct {
	ID                int    `db:"id"`
	Login             string `db:"login"`
	EncryptedPassword string `db:"encrypted_password"`
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type Order struct {
	ID          int          `json:"-" db:"id"`
	OrderNumber string       `json:"number" db:"order_number"`
	UserID      int          `json:"-" db:"user_id"`
	Status      string       `json:"status" db:"status"`
	Accrual     *NullFloat64 `json:"accrual,omitempty" db:"accrual"`
	UploadedAt  string       `json:"uploaded_at" db:"uploaded_at"`
}

type Withdraw struct {
	ID          int     `json:"-" db:"id"`
	OrderNumber string  `json:"order" db:"order_number"`
	Sum         float64 `json:"sum" db:"sum"`
	ProcessedAt string  `json:"processed_at" db:"processed_at"`
}

type JSONBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type JSONWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type AccrualOrderState struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// NullFloat64 is an alias for sql.NullInt64 data type
type NullFloat64 sql.NullFloat64

// Scan implements the Scanner interface for NullFloat64
func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}
	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{f.Float64, false}
	} else {
		*nf = NullFloat64{f.Float64, true}
	}
	return nil
}

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}
