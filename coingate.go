package coingate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"strconv"
	"strings"
	"time"
)

// Main Coingate class
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
		Set("Access-Signature", c.getHMACSignature(nonce))

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

// Fields for creating order request
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

// GET parameters for ListOrders request
type ListOrdersRequest struct {
	PerPage int    `json:"per_page"`
	Page    int    `json:"page"`
	Sort    string `json:"sort"`
}

// The order
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

// List of orders
type Orders struct {
	CurrentPage int     `json:"current_page"`
	PerPage     int     `json:"per_page"`
	TotalOrders int     `json:"total_orders"`
	TotalPages  int     `json:"total_pages"`
	Orders      []Order `json:"orders"`
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
