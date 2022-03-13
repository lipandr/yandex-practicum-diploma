package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lipandr/yandex-practicum-diploma/internal/dao"
	"github.com/lipandr/yandex-practicum-diploma/internal/types"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net/http"
	"time"
)

const tooManyRequestTemplate = "No more than %d requests per minute allowed"

type AccrualProcessor interface {
	GetOrderStatus(orderID string) *types.AccrualOrderState
	Run()
}

type accrualProcessor struct {
	address   string
	rateLimit int
	poolSize  int
	limiter   *rate.Limiter
	dao       *dao.DAO

	OrderQueue chan string
}

func (a *accrualProcessor) Run() {
	go func() {
		for {
			orderList, err := a.dao.GetOrdersForProcessing()
			if err != nil {
				time.Sleep(time.Minute)
				continue
			}
			for _, orderID := range orderList {
				a.OrderQueue <- orderID
			}
		}
	}()

}

func (a *accrualProcessor) GetOrderStatus(orderID string) *types.AccrualOrderState {
	res, err := http.Get(fmt.Sprintf("%s/api/orders/%s", a.address, orderID))
	if err != nil {
		//log.Println(err)
		return nil
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode == http.StatusTooManyRequests {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			//log.Println(err)
			return nil
		}
		var rl int
		_, err = fmt.Sscanf(tooManyRequestTemplate, string(resBody), &rl)
		if err != nil {
			//log.Println(err)
			return nil
		}
		a.setLimit(rl)
	}

	if res.StatusCode != http.StatusOK {
		return nil
	}

	var aos types.AccrualOrderState
	if err := json.NewDecoder(res.Body).Decode(&aos); err != nil {
		//log.Println(err)
		return nil
	}

	return &aos
}

func NewAccrualProcessor(dao *dao.DAO, addr string, poolSize int) AccrualProcessor {
	ap := &accrualProcessor{
		dao:        dao,
		address:    addr,
		poolSize:   poolSize,
		OrderQueue: make(chan string, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		go ap.queueWorker()
	}

	return ap
}

func (a *accrualProcessor) setLimit(n int) {
	if n <= 0 {
		a.limiter = nil
		return
	}
	a.limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(n)), n)
}

func (a *accrualProcessor) queueWorker() {
	for orderID := range a.OrderQueue {
		if a.limiter != nil && !a.limiter.Allow() {
			err := a.limiter.Wait(context.Background())
			if err != nil {
				//log.Println(err)
				return
			}
		}
		orderStatus := a.GetOrderStatus(orderID)
		if orderStatus != nil {
			if err := a.dao.UpdateOrderState(orderStatus); err != nil {
				log.Println(err)
			}
		}
	}
}
