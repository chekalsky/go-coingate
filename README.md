# Coingate.com Go API client

[![GoDoc](https://godoc.org/github.com/chekalskiy/go-coingate?status.svg)](https://godoc.org/github.com/chekalskiy/go-coingate)

Usage is very simple:

```golang
package main

import (
	"log"
	"github.com/chekalskiy/go-coingate"
)

func main() {
    Coingate := coingate.New(1234, "Key", "Secret", false)
    
    order, err := Coingate.CreateOrder(coingate.OrderRequest{
        Price: "10.00",
        Currency: "USD",
        ReceiveCurrency: "BTC",
    })
    
    if err != nil {
        log.Fatalln(err)
    }
    
    log.Println(order)
}
```

Callbacks are in progress...