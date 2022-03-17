package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/joeljunstrom/go-luhn"
	"github.com/lipandr/yandex-practicum-diploma/internal/types"
)

// UserRegistration Handler регистрация и аутентификация нового пользователя.
func (a *application) UserRegistration(w http.ResponseWriter, r *http.Request) {
	var user types.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := a.svc.UserRegistration(&user)
	if err != nil {
		if errors.Is(err, types.ErrUsersAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.Token))

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// UserAuthentication Handler аутентификация зарегистрированного пользователя.
func (a *application) UserAuthentication(w http.ResponseWriter, r *http.Request) {
	var user types.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := a.svc.UserAuthentication(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.Token))

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// ReceiveOrder Handler принятие в обработку нового заказа.
func (a *application) ReceiveOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserID).(int)

	value, err := ioutil.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()

	if err != nil || len(value) == 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orderNumber := string(value)
	if err := ValidateOrderNumber(orderNumber); err != nil {
		http.Error(w, types.ErrOrderNumberInvalid.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.svc.ReceiveOrder(userID, orderNumber); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, types.ErrOrderUploadedByUser) {
			status = http.StatusOK
		}
		if errors.Is(err, types.ErrOrderUploadedByOtherUser) {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders Handler получение списка загруженных заказов для начисления.
func (a *application) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserID).(int)

	orders, err := a.svc.GetOrders(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, errors.New("order list is empty").Error(), http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetBalance Handler получение текущего баланса пользователя.
func (a *application) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserID).(int)

	crnt, wthd, err := a.svc.GetBalance(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	res := types.JSONBalance{
		Current:   crnt,
		Withdrawn: wthd,
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// WithdrawRequest Handler запрос на списание начислений.
func (a *application) WithdrawRequest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserID).(int)

	var req types.JSONWithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ValidateOrderNumber(req.Order); err != nil {
		http.Error(w, types.ErrOrderNumberInvalid.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.svc.WithdrawRequest(userID, req.Order, req.Sum); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, types.ErrInsufficientAccruals) {
			status = http.StatusPaymentRequired
		}
		w.WriteHeader(status)
		return
	}
}

// GetWithdrawals Handler получение списка списаний начислений.
func (a *application) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserID).(int)

	wthd, err := a.svc.GetWithdrawals(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(wthd) == 0 {
		http.Error(w, errors.New("the list is empty").Error(), http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(wthd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ValidateOrderNumber(orderID string) error {
	if ok := luhn.Valid(orderID); !ok {
		return types.ErrOrderNumberInvalid
	}
	if len(orderID) < 5 || len(orderID) > 15 {
		return types.ErrOrderNumberInvalid
	}
	return nil
}
