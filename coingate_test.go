package coingate_test

import (
	"github.com/chekalskiy/go-coingate"
	"reflect"
	"testing"
)

const (
	AppId     = 1015
	ApiKey    = "WV5FGZRuoCstBiHayYXbp3"
	ApiSecret = "ni6FmIUTxBSR8PK2JeDA5NVjpzc4tqZO"
)

var Coingate *coingate.Coingate
var BetweenTests struct {
	Order coingate.Order
}

func init() {
	Coingate = coingate.New(AppId, ApiKey, ApiSecret, true)
}

func TestCoingate_Ping(t *testing.T) {
	res := Coingate.Ping()

	if res != true {
		t.Fatal("Ping doesn't responeded")
	}
}

func TestCoingate_CreateOrder(t *testing.T) {
	price := "10.01"

	o, err := Coingate.CreateOrder(coingate.OrderRequest{
		Price:           price,
		Currency:        "USD",
		ReceiveCurrency: "BTC",
	})

	if err != nil {
		t.Fatal("Error on create order")
	}

	if o.Id <= 0 {
		t.Fatal("Wrong response from server")
	}

	if o.Price != price {
		t.Fatal("Wrong price in created order")
	}

	if len(o.PaymentUrl) == 0 {
		t.Fatal("Empty payment URL")
	}

	BetweenTests.Order = o
}

func TestCoingate_GetOrder(t *testing.T) {
	o, err := Coingate.GetOrder(BetweenTests.Order.Id)

	if err != nil {
		t.Fatal("Error on getting order")
	}

	if !reflect.DeepEqual(BetweenTests.Order, o) {
		t.Fatal("Order doesn't match created")
	}
}

func TestCoingate_ListOrders(t *testing.T) {
	o, err := Coingate.ListOrders(coingate.ListOrdersRequest{})

	if err != nil {
		t.Fatal("Error on getting orders")
	}

	if !reflect.DeepEqual(BetweenTests.Order, o.Orders[0]) {
		t.Fatal("Order is not found in list")
	}

	o, err = Coingate.ListOrders(coingate.ListOrdersRequest{
		Page:    -10,
		PerPage: 0,
		Sort:    "bad_sort",
	})

	if err != nil {
		t.Fatal("Error on getting orders")
	}

	if !reflect.DeepEqual(BetweenTests.Order, o.Orders[0]) {
		t.Fatal("Order is not found in list")
	}
}

func TestCoingate_WithErrors(t *testing.T) {
	_, err := Coingate.CreateOrder(coingate.OrderRequest{})

	if err == nil {
		t.Fatal("Should be an error")
	}

	_, err = Coingate.GetOrder(0)

	if err == nil {
		t.Fatal("Should be an error")
	}

	BadCoingate := coingate.New(0, "", "", true)
	_, err = BadCoingate.ListOrders(coingate.ListOrdersRequest{})

	if err == nil {
		t.Fatal("Should be an error")
	}
}
