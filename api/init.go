package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func Init(cfgname string) (res *Database, err error) {
	var cfg map[string]interface{}
	JsonFile, err := os.Open(cfgname + ".json")
	if err != nil {
		return nil, err
	}
	defer JsonFile.Close()
	bytes, err := io.ReadAll(JsonFile)
	json.Unmarshal(bytes, &cfg)
	log.Print("Got config with following values:")
	for key, value := range cfg {
		log.Println(key, ":", value)
	}
	var db Database
	db.Time = timeConv(time.Now())
	resdb, err := sql.Open("postgres", getConnectionString(cfg))
	if err != nil {
		return nil, err
	}
	db.DB = resdb
	log.Printf("time converted: %d", db.Time)
	_, err = http.Get(source + "ping")
	if err != nil {
		log.Printf("Binance inreachable %s", err.Error())
		return nil, err
	}
	return &db, nil
}
