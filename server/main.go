package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	UrlExchangeRate = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ReqTimeout      = time.Millisecond * 200
	DbTimeout       = time.Millisecond * 10
)

var db *sql.DB

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"var_bid"`
	PctChange  string `json:"pct_change"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ExchangeRate struct {
	Usdbrl Usdbrl
}

type dollarResponse struct {
	Bid string `json:"bid"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	createDb()
	http.HandleFunc("/cotacao", handleDollarExchange)
	http.ListenAndServe(":8080", nil)
}

func handleDollarExchange(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), ReqTimeout)
	defer cancel()
	exchangeRate, err := requestExchangeRate(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Dolar:%s", exchangeRate.Usdbrl.Bid)
	insertExchangeRate(ctx, db, &exchangeRate.Usdbrl)
	json.NewEncoder(w).Encode(dollarResponse{
		Bid: exchangeRate.Usdbrl.Bid,
	})
}

func requestExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", UrlExchangeRate, nil)
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

func insertExchangeRate(ctx context.Context, db *sql.DB, usdBrl *Usdbrl) error {
	ctx, cancel := context.WithTimeout(ctx, DbTimeout)
	defer cancel()
	stmt, err := db.Prepare("insert into usdbrl(code,codein,name,high,low,var_bid,pct_change,bid,ask,timestamp,create_date) " +
		"values (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(
		ctx,
		usdBrl.Code,
		usdBrl.Codein,
		usdBrl.Name,
		usdBrl.High,
		usdBrl.Low,
		usdBrl.VarBid,
		usdBrl.PctChange,
		usdBrl.Bid,
		usdBrl.Ask,
		usdBrl.Timestamp,
		usdBrl.CreateDate)
	if err != nil {
		return err
	}
	return nil
}

func createDb() error {
	sts := `
CREATE TABLE IF NOT EXISTS usdbrl(id INTEGER PRIMARY KEY, 
					code TEXT,
					codein TEXT,
					name TEXT,
					high TEXT,
					low TEXT,
					var_bid TEXT,
					pct_change TEXT,
					bid TEXT,
					ask TEXT,
					timestamp TEXT,
					create_date TEXT);
`
	_, err := db.Exec(sts)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("table cars created")
	return nil
}
