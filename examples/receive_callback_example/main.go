package main

import (
	"fmt"
	"github.com/chekalskiy/go-coingate"
	"log"
	"net/http"
)

var cg *coingate.Coingate

const (
	AppID     = 1015
	APIKey    = "WV5FGZRuoCstBiHayYXbp3"
	APISecret = "ni6FmIUTxBSR8PK2JeDA5NVjpzc4tqZO"
)

// After running main.go execute:
// `curl -X POST -d "id=343&order_id=ORDER-1415020039&status=paid&price=1050.99&currency=USD&receive_currency=EUR&receive_amount=926.73&btc_amount=4.81849315&created_at=2014-11-03T13:07:28%2B03:00" http://localhost:8000/payments/callback`
func main() {
	cg = coingate.New(AppID, APIKey, APISecret, true)

	http.HandleFunc("/payments/callback", receiveCallback)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func receiveCallback(w http.ResponseWriter, r *http.Request) {
	p, err := cg.ProcessCallback(r)

	if err == nil {
		log.Println("Order ID:", p.ID)
		log.Println("Order status:", p.Status)
	} else {
		log.Fatalln("Error", err)
	}

	fmt.Fprintf(w, "ok")
}
