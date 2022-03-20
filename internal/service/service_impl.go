package service

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"golang.org/x/crypto/bcrypt"

	"github.com/lipandr/yandex-practicum-diploma/internal/types"
)

// UserRegistration метод Service регистрации нового пользователя.
func (svc *service) UserRegistration(user *types.UserRequest) (*types.AuthResponse, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	encPass := string(b)
	_, err = svc.dao.NewUser(user.Login, encPass)
	if err != nil {
		return nil, types.ErrUsersAlreadyExists
	}
	return svc.UserAuthentication(user)
}

// UserAuthentication метод Service аутентификации существующего пользователя.
func (svc *service) UserAuthentication(user *types.UserRequest) (*types.AuthResponse, error) {
	u, err := svc.dao.GetUserByLogin(user.Login)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(user.Password))
	if err != nil {
		return nil, types.ErrUsersNotAuthenticated
	}
	token, err := svc.generateToken(64)
	if err != nil {
		return nil, err
	}
	err = svc.dao.SaveToken(u.ID, token)
	if err != nil {
		return nil, err
	}
	return &types.AuthResponse{
		Token: token,
	}, nil
}

// GetUserIDByToken метод Service получения пользователя по токену.
func (svc *service) GetUserIDByToken(token string) (int, error) {
	return svc.dao.GetToken(token)
}

// ReceiveOrder метод Service добавления нового заказа для расчета начислений.
func (svc *service) ReceiveOrder(userID int, orderNumber string) error {
	// проверка - были ли списания по заказу
	if err := svc.dao.IsOrderWithdrawn(orderNumber); err != nil {
		return err
	}
	// проверка - был загружен ранее
	if err := svc.dao.IsOrderExists(userID, orderNumber); err != nil {
		return err
	}
	if err := svc.dao.NewOrder(userID, orderNumber); err != nil {
		return err
	}
	return nil
}

// GetOrders метод Service получения списка заказов для расчета начислений пользователя.
func (svc *service) GetOrders(userID int) ([]types.Order, error) {
	orders, err := svc.dao.GetOrderList(userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetBalance метод Service получения баланса начислений пользователя.
func (svc *service) GetBalance(userID int) (float64, float64, error) {
	a, err := svc.dao.GetAccruals(userID)
	if err != nil {
		return 0, 0, err
	}
	w, err := svc.dao.GetTotalWithdrawals(userID)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	b := math.Round((a-w)*100) / 100
	return b, w, nil
}

// WithdrawRequest метод Service запроса на списание начислений.
func (svc *service) WithdrawRequest(userID int, orderNumber string, sum float64) error {
	// проверка - производилось ли списание по заказу ранее
	if err := svc.dao.IsOrderWithdrawn(orderNumber); err != nil {
		return err
	}
	b, _, err := svc.GetBalance(userID)
	if err != nil {
		return err
	}
	if b < sum {
		return types.ErrInsufficientAccruals
	}
	if err = svc.dao.NewWithdrawal(userID, sum, orderNumber); err != nil {
		return err
	}
	return nil
}

// GetWithdrawals метод Service получения списка списаний пользователя.
func (svc *service) GetWithdrawals(userID int) ([]types.Withdraw, error) {
	res, err := svc.dao.GetWithdrawalsList(userID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// generateToken метод Service генерации случайного токена для авторизации пользователя.
func (svc *service) generateToken(n int) (string, error) {
	const letters = "zxcvbnmasdfghjklqwertyuiop1234567890"
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		res[i] = letters[num.Int64()]
	}
	return string(res), nil
}
