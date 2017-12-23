package coingate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Coingate main class
type Coingate struct {
	appID     int
	apiKey    string
	apiSecret string
	baseURL   string
}

var urlLive = "api.coingate.com/v1"
var urlSandbox = "api-sandbox.coingate.com/v1"

// New creates a new Coingate instance
//
// It requires app data which you can get on https://coingate.com
func New(appID int, apiKey string, apiSecret string, isSandbox bool) *Coingate {
	u := urlLive
	if isSandbox {
		u = urlSandbox
	}

	c := &Coingate{
		appID:     appID,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   u,
	}

	return c
}

// CreateOrder creates new order with specified params
func (c *Coingate) CreateOrder(d CreateOrderRequest) (Order, error) {
	res, err := c.request("POST", "/orders", d)

	if err != nil {
		return Order{}, err
	}

	var o Order
	json.Unmarshal([]byte(res), &o)

	return o, nil
}

// GetOrder gets information about order by it's id
func (c *Coingate) GetOrder(id int) (Order, error) {
	res, err := c.request("GET", fmt.Sprintf("/orders/%d", id), nil)

	if err != nil {
		return Order{}, err
	}

	var o Order
	json.Unmarshal([]byte(res), &o)

	return o, nil
}

// ListOrders gets list of orders
func (c *Coingate) ListOrders(d ListOrdersRequest) (Orders, error) {
	if d.PerPage <= 0 {
		d.PerPage = 10
	}

	if d.Page < 1 {
		d.Page = 1
	}

	if len(d.Sort) == 0 {
		d.Sort = "created_at_desc"
	}

	res, err := c.request("GET", "/orders", d)

	if err != nil {
		return Orders{}, err
	}

	var o Orders
	json.Unmarshal([]byte(res), &o)

	return o, nil
}

// Ping requests the ping endpoint
func (c *Coingate) Ping() bool {
	res, err := c.request("GET", "/ping", nil)

	if err != nil {
		return false
	}

	var p pong
	json.Unmarshal([]byte(res), &p)

	if p.Ping == "pong" {
		return true
	}

	return false
}

// ProcessCallback gets a http.Request and parses POST fields to struct CallbackData
func (c *Coingate) ProcessCallback(r *http.Request) (CallbackData, error) {
	err := r.ParseForm()

	if err != nil {
		return CallbackData{}, fmt.Errorf("Form process error: %s", err)
	}

	var cd CallbackData
	cd.ID, _ = strconv.Atoi(r.Form.Get("id"))
	cd.OrderID = r.Form.Get("order_id")
	cd.Status = r.Form.Get("status")
	cd.Price = r.Form.Get("price")
	cd.Currency = r.Form.Get("currency")
	cd.ReceiveCurrency = r.Form.Get("receive_currency")
	cd.ReceiveAmount = r.Form.Get("receive_amount")
	cd.BtcAmount = r.Form.Get("btc_amount")
	cd.CreatedAt, _ = time.Parse("2006-01-02T15:04:05-07:00", r.Form.Get("created_at"))

	return cd, nil
}

func (c *Coingate) request(method string, uri string, data interface{}) (string, error) {
	nonce := time.Now().UnixNano()
	request := gorequest.New()

	url := fmt.Sprintf("https://%s/%s", c.baseURL, strings.Trim(uri, "/"))

	request.CustomMethod(method, url)

	if method == gorequest.POST {
		request.Type("multipart").
			Send(data)
	} else if method == gorequest.GET {
		request.Query(data)
	}

	request.Set("Accept", "application/json").
		Set("Access-Nonce", strconv.Itoa(int(nonce))).
		Set("Access-Key", c.apiKey).
		Set("Access-Signature", c.getHMACSignature(nonce)).
		Timeout(time.Second * 15)

	resp, body, errs := request.End()

	if len(errs) > 0 {
		return "", errs[0]
	}

	if resp.StatusCode != 200 {
		var m errorResponse
		json.Unmarshal([]byte(body), &m)

		// Because sometimes there is Reason + Message, sometimes only Error
		return "", fmt.Errorf("Error %d: %s %s%s", resp.StatusCode, m.Reason, m.Message, m.Error)
	}

	return body, nil
}

func (c *Coingate) getHMACSignature(nonce int64) string {
	m := fmt.Sprintf("%d%d%s", nonce, c.appID, c.apiKey)

	secret := []byte(c.apiSecret)
	message := []byte(m)

	hash := hmac.New(sha256.New, secret)
	hash.Write(message)

	return hex.EncodeToString(hash.Sum(nil))
}

// CreateOrderRequest is a struct with fields for creating order request
type CreateOrderRequest struct {
	OrderID         string `json:"order_id"`
	Price           string `json:"price"`
	Currency        string `json:"currency"`
	ReceiveCurrency string `json:"receive_currency"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	CallbackURL     string `json:"callback_url"`
	CancelURL       string `json:"cancel_url"`
	SuccessURL      string `json:"success_url"`
}

// ListOrdersRequest is a struct with GET parameters for ListOrders request
type ListOrdersRequest struct {
	PerPage int    `json:"per_page"`
	Page    int    `json:"page"`
	Sort    string `json:"sort"`
}

// Order is The Order
type Order struct {
	ID             int       `json:"id"`
	Currency       string    `json:"currency"`
	BitcoinURI     string    `json:"bitcoin_uri"`
	Status         string    `json:"status"`
	Price          string    `json:"price"`
	BtcAmount      string    `json:"btc_amount"`
	CreatedAt      time.Time `json:"created_at"`
	ExpireAt       time.Time `json:"expire_at"`
	BitcoinAddress string    `json:"bitcoin_address"`
	OrderID        string    `json:"order_id"`
	PaymentURL     string    `json:"payment_url"`
}

// Orders is a list of orders
type Orders struct {
	CurrentPage int     `json:"current_page"`
	PerPage     int     `json:"per_page"`
	TotalOrders int     `json:"total_orders"`
	TotalPages  int     `json:"total_pages"`
	Orders      []Order `json:"orders"`
}

// CallbackData is a struct of POST fields which we receive as a callback from Coingate
type CallbackData struct {
	ID              int       `json:"id"`
	OrderID         string    `json:"order_id"`
	Status          string    `json:"status"`
	Price           string    `json:"price"`
	Currency        string    `json:"currency"`
	ReceiveCurrency string    `json:"receive_currency"`
	ReceiveAmount   string    `json:"receive_amount"`
	BtcAmount       string    `json:"btc_amount"`
	CreatedAt       time.Time `json:"created_at"`
}

type pong struct {
	Ping string    `json:"ping"`
	Time time.Time `json:"time"`
}

type errorResponse struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}
