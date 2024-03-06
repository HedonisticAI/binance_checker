package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1s&startTime=1609633762381&limit=1
const (
	source = "http://api.binance.com/api/v3/"
)

func (db *Database) GetInfo(c *gin.Context) {

	var old float64
	var new float64
	defer db.Wg.Done()
	db.Wg.Add(1)
	var req GetInfo
	req.Ticker = c.Request.URL.Query().Get("Ticker")
	req.From, _ = time.Parse("01-JAN-2006 15:04:00", c.Request.URL.Query().Get("date_from"))
	req.To, _ = time.Parse("01-JAN-2006 15:04:00", c.Request.URL.Query().Get("date_to"))
	if req.Ticker == "" {
		log.Printf("Request with no symbol")
		c.JSON(http.StatusBadRequest, "Request with no symbol")
		return

	}
	log.Printf("Got request with %s ticker from %v to %v", req.Ticker, req.From, req.To)
	if req.To.Before(req.From) {
		c.JSON(http.StatusBadRequest, "Time From > Time To")
		return
	}

	if timeConv(req.From) < db.Time {
		db.OldInfoRequest(req.From, req.Ticker)
		res, err := db.DB.Query("SELECT price FROM prices WHERE pair=$1 AND price_time=$2", req.Ticker, timeConv(req.From))
		if err != nil {
			db.OldInfoRequest(req.From, req.Ticker)
			c.JSON(http.StatusOK, "no price found")
		}
		log.Printf("Performing search and request for old info with %s %v", req.Ticker, req.From)
		res.Scan(&old)
		if old == 0.0 {
			res, err := db.DB.Query("SELECT price FROM prices WHERE pair=$1 LIMIT 1", req.Ticker, timeConv(req.From))
			if err != nil {
				db.OldInfoRequest(req.From, req.Ticker)
				c.JSON(http.StatusOK, "no price found")
				res.Scan(&old)
			}
		}

	} else {
		log.Printf("Performing search and request for current info with %s %v", req.Ticker, req.From)
		res, _ := db.DB.Query("SELECT price FROM prices WHERE pair=$1 AND  price_time=$2", req.Ticker, timeConv(req.From))
		res.Scan(&old)

	}
	if timeConv(req.To) < db.Time {
		res, err := db.DB.Query("SELECT price FROM prices WHERE pair=$1 AND  price_time=$2", req.Ticker, timeConv(req.To))
		if err != nil {
			db.OldInfoRequest(req.To, req.Ticker)
			c.JSON(http.StatusOK, "internal error while getting price")
		}
		log.Printf("Performing search and request for old info with %s %v", req.Ticker, req.To)
		res.Scan(&new)
		if new == 0.0 {
			res, err := db.DB.Query("SELECT price FROM prices WHERE pair=$1 LIMIT 1", req.Ticker, timeConv(req.From))
			if err != nil {
				db.OldInfoRequest(req.From, req.Ticker)
				c.JSON(http.StatusOK, "no price found")
				res.Scan(&new)
			}
		}

	} else {

		log.Printf("Performing search and request for current info with %s %v", req.Ticker, req.To)
		res, _ := db.DB.Query("SELECT price FROM prices WHERE pair=$1 AND  price_time=$2", req.Ticker, timeConv(req.From))
		res.Scan(&new)

	}
	resp := formResponse(req.Ticker, old, new)
	c.JSON(http.StatusOK, resp)
}

func (db *Database) AddPair(c *gin.Context) {
	var req AddPair
	defer db.Wg.Done()
	db.Wg.Add(1)
	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil || bytes == nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	log.Printf("Got AddPair request")
	json.Unmarshal(bytes, &req)
	if validPair(req.Pair) {
		_, err := db.DB.Exec("INSERT INTO pairs(pair) VALUES ($1)", req.Pair)
		if err != nil {
			log.Printf("Error inserting pair %s", err.Error())
			c.JSON(http.StatusOK, "SQL error"+err.Error())
		} else {
			c.JSON(http.StatusOK, "Pair added successfully")
		}
	} else {
		c.JSON(http.StatusBadRequest, "Pair invalid or not found on Binance")
	}

}

func (db *Database) Fill() error {
	var cfg map[string]interface{}
	for {
		t1 := time.Now()
		db.Wg.Wait()
		rows, err := db.DB.Query("SELECT pair FROM pairs")
		if err == nil {
			for rows.Next() {
				var pair string
				err := rows.Scan(&pair)
				if err == nil {
					res, err := http.Get(source + "ticker/price?symbol=" + pair)
					if err != nil {
						log.Printf("error while filling")
					}
					bytes, _ := io.ReadAll(res.Body)
					json.Unmarshal(bytes, &cfg)
					price, _ := strconv.ParseFloat(cfg["price"].(string), 64)
					_, err = db.DB.Exec("INSERT INTO prices (pair, price, price_time) VALUES ($1,$2,$3)", pair, price, timeConv(t1))
					if err != nil {
						log.Printf("error while filling")
					}
					log.Printf("Gor price of %s by time%d %f", pair, timeConv(t1), price)
				} else {
					log.Printf("Error during filling DB: %s", err.Error())
				}
			}

		} else {
			log.Print(err.Error())
		}
		time.Sleep(1*time.Minute - time.Since(t1))
	}
}

func (db *Database) OldInfoRequest(time time.Time, pair string) error {
	res, err := http.Get(formOldReqString(pair, time))
	log.Print(formOldReqString(pair, time))
	if err != nil {
		log.Printf("Error making old info request: %s", err.Error())
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("Here")
	f, err := dataParse(bytes)
	if err != nil {
		return err
	}
	log.Printf("received data %d", f.time)
	final, err := strconv.ParseFloat(f.price, 64)
	if err != nil {
		return err
	}
	db.DB.Exec("INSERT INTO prices (pair, price,  price_time) VALUES ($1,$2,3$)", pair, final, f.time)
	return nil
}
