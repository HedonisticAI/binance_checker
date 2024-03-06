package api

import (
	"database/sql"
	"sync"
	"time"
)

type Database struct {
	Wg   sync.WaitGroup
	DB   *sql.DB
	Time int64
}

type AddPair struct {
	Pair string `json:"ticker"`
}
type GetInfo struct {
	Ticker string
	From   time.Time
	To     time.Time
}

type Response struct {
	Ticker string  `json:"Ticker"`
	Price  float64 `json:"Price"`
	Diff   string  `json:"Difference"`
}

type OldInfo struct {
	price string
	time  int64
}
