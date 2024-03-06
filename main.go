package main

import (
	"gkhazizov_test/api"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	t := time.Now().Add(time.Duration(-10) * time.Minute)
	db, err := api.Init("config")
	if err != nil {
		log.Println("Error initializing database or connecting to binance:", err.Error())
		return
	}
	go db.Fill()

	defer db.DB.Close()
	r := gin.Default()
	r.GET("/fetch", db.GetInfo)
	r.POST("/add_ticker", db.AddPair)
	db.OldInfoRequest(t, "BTCUSDT")
	r.Run()
}
