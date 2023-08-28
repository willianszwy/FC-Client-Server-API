package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ServerUrl = "http://localhost:8080/cotacao"
	timeout   = time.Millisecond * 300
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func main() {
	exchangeRate, err := requestExchangeRate()
	if err != nil {
		panic(err)
	}
	log.Println("Dollar exchange rate", exchangeRate.Bid)
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("Dolar:%s", exchangeRate.Bid))
	if err != nil {
		panic(err)
	}
}

func requestExchangeRate() (*ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", ServerUrl, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	var exchangeRate ExchangeRate
	err = json.Unmarshal(body, &exchangeRate)
	if err != nil {
		return nil, err
	}
	return &exchangeRate, nil
}
