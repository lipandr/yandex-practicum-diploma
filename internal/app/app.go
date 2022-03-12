package app

import (
	"github.com/gorilla/mux"
	"github.com/lipandr/yandex-practicum-diploma/internal/config"
	"github.com/lipandr/yandex-practicum-diploma/internal/service"
	"net/http"
)

// Application интерфейс приложения
type Application interface {
	Run() error
	UserRegistration(w http.ResponseWriter, r *http.Request)
	UserAuthentication(w http.ResponseWriter, r *http.Request)
	ReceiveOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
	WithdrawRequest(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

type application struct {
	cfg config.Config
	svc service.Service
}

// NewApp метод конструктор приложения
func NewApp(cfg config.Config, svc service.Service) Application {
	return &application{
		cfg: cfg,
		svc: svc,
	}
}

// Run запуск сервера приложения
func (a *application) Run() error {
	r := mux.NewRouter()

	r.Use(GzipMiddleware, AuthMiddleware(a.svc))

	r.HandleFunc("/api/user/register", a.UserRegistration).Methods(http.MethodPost)
	r.HandleFunc("/api/user/login", a.UserAuthentication).Methods(http.MethodPost)

	r.HandleFunc("/api/user/orders", a.ReceiveOrder).Methods(http.MethodPost)
	r.HandleFunc("/api/user/orders", a.GetOrders).Methods(http.MethodGet)
	r.HandleFunc("/api/user/balance", a.GetBalance).Methods(http.MethodGet)
	r.HandleFunc("/api/user/balance/withdraw", a.WithdrawRequest).Methods(http.MethodPost)
	r.HandleFunc("/api/user/withdrawals", a.GetWithdrawals).Methods(http.MethodGet)

	return http.ListenAndServe(a.cfg.RunAddress, r)
}
