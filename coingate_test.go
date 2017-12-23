package coingate_test

import (
	"fmt"
	"github.com/chekalskiy/go-coingate"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	AppID     = 1015
	APIKey    = "WV5FGZRuoCstBiHayYXbp3"
	APISecret = "ni6FmIUTxBSR8PK2JeDA5NVjpzc4tqZO"
)

var Coingate *coingate.Coingate
var BetweenTests struct {
	Order coingate.Order
}

func init() {
	Coingate = coingate.New(AppID, APIKey, APISecret, true)
}

func TestCoingate_Ping(t *testing.T) {
	res := Coingate.Ping()

	if res != true {
		t.Fatal("Ping doesn't responeded")
	}
}

func TestCoingate_CreateOrder(t *testing.T) {
	price := "10.01"

	o, err := Coingate.CreateOrder(coingate.CreateOrderRequest{
		Price:           price,
		Currency:        "USD",
		ReceiveCurrency: "BTC",
	})

	if err != nil {
		t.Fatal("Error on create order")
	}

	if o.ID <= 0 {
		t.Fatal("Wrong response from server")
	}

	if o.Price != price {
		t.Fatal("Wrong price in created order")
	}

	if len(o.PaymentURL) == 0 {
		t.Fatal("Empty payment URL")
	}

	BetweenTests.Order = o
}

func TestCoingate_GetOrder(t *testing.T) {
	o, err := Coingate.GetOrder(BetweenTests.Order.ID)

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

func TestCoingate_ProcessCallback(t *testing.T) {
	testTime := time.Now()
	testParams := fmt.Sprintf(`id=12345&order_id=ORDER-1415020039&status=paid&price=1050.99&currency=USD&receive_currency=EUR&receive_amount=926.73&btc_amount=4.849315&created_at=%s`, url.QueryEscape(testTime.Format("2006-01-02T15:04:05-07:00")))

	r, err := http.NewRequest("POST", "", strings.NewReader(testParams))

	if err != nil {
		t.Fatal("Request creating error", err)
	}

	r.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	cd, err := Coingate.ProcessCallback(r)

	if err != nil {
		t.Fatal("Error processing callback", err)
	}

	if cd.ID != 12345 {
		t.Fatal("Wrong callback object ID")
	}

	if cd.CreatedAt.Format(time.RFC3339) != testTime.Format(time.RFC3339) {
		t.Fatal("Wrong callback object time")
	}

	// ---------

	r, err = http.NewRequest("POST", "", nil)

	if err != nil {
		t.Fatal("Request creating error", err)
	}

	cd, err = Coingate.ProcessCallback(r)

	if err == nil {
		t.Fatal("Should be an error")
	}

	// ---------

	r, err = http.NewRequest("POST", "", strings.NewReader("bad=1"))

	if err != nil {
		t.Fatal("Request creating error", err)
	}

	r.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	cd, _ = Coingate.ProcessCallback(r)

	if cd.ID > 0 {
		t.Fatal("Request creating error", err)
	}
}

func TestCoingate_WithErrors(t *testing.T) {
	_, err := Coingate.CreateOrder(coingate.CreateOrderRequest{})

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
