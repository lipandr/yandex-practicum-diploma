package service

import (
	"github.com/lipandr/yandex-practicum-diploma/internal/dao"
	"github.com/lipandr/yandex-practicum-diploma/internal/types"
)

type Service interface {
	UserRegistration(user *types.UserRequest) (*types.AuthResponse, error)
	UserAuthentication(user *types.UserRequest) (*types.AuthResponse, error)
	ReceiveOrder(userID int, orderNumber string) error
	GetOrders(userID int) ([]types.Order, error)
	GetBalance(userID int) (float64, float64, error)
	WithdrawRequest(userID int, order string, sum float64) error
	GetWithdrawals(userID int) ([]types.Withdraw, error)
	GetUserIDByToken(token string) (int, error)
}

type service struct {
	dao *dao.DAO
}

func NewService(dao *dao.DAO) (*service, error) {
	return &service{
		dao: dao,
	}, nil
}
