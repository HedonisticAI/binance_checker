package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func getConnectionString(cfg map[string]interface{}) string {
	res := fmt.Sprintf("host=%s user=%s port=%d dbname=%s password=%s sslmode=disable",
		cfg["host"].(string), cfg["user"].(string), int64(cfg["port"].(float64)), cfg["name"].(string), cfg["password"].(string))
	log.Println("Generated connection string:", res)

	return res
}

func timeConv(t time.Time) int64 {
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
	return tUnixMilli
}

func persentDiff(a, b float64) float64 {
	res := (a - b) / b
	return res * 100.0
}

func validPair(pair string) bool {
	req, err := http.Get(source + "avgPrice?symbol=" + pair)
	log.Printf("Checking symbol validity %s", pair)
	if req.StatusCode == 200 && err == nil {
		log.Printf("Symbol %s valid", pair)
		return true

	}
	return false
}

func formOldReqString(pair string, time time.Time) string {
	return source + fmt.Sprintf("klines?symbol=%s&startTime=%d&limit=1&linterval=1s", pair, timeConv(time))

}
func formResponse(pair string, old float64, new float64) Response {
	var res Response
	res.Ticker = pair
	res.Price = new
	res.Diff = fmt.Sprintf("%f", persentDiff(new, old)) + "%"
	return res
}

func dataParse(p []byte) (*OldInfo, error) {
	var tmp []json.RawMessage
	var final OldInfo

	if err := json.Unmarshal(p, &tmp); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tmp[0], &final.time); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tmp[4], &final.price); err != nil {
		return nil, err
	}
	return &final, nil
}
