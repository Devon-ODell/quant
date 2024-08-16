package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type OHLC struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func fetchHistoricalData(pair string, interval int, start, end time.Time) ([]OHLC, error) {
	url := fmt.Sprintf("https://api.kraken.com/0/public/OHLC?pair=%s&interval=%d&since=%d",
		pair, interval, start.Unix())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	data := result["result"].(map[string]interface{})[pair].([]interface{})
	var ohlcData []OHLC

	for _, v := range data {
		item := v.([]interface{})
		timestamp, _ := strconv.ParseInt(item[0].(string), 10, 64)
		if timestamp > end.Unix() {
			break
		}
		open, _ := strconv.ParseFloat(item[1].(string), 64)
		high, _ := strconv.ParseFloat(item[2].(string), 64)
		low, _ := strconv.ParseFloat(item[3].(string), 64)
		close, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[6].(string), 64)

		ohlcData = append(ohlcData, OHLC{
			Time:   timestamp,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return ohlcData, nil
}

func saveToCSV(data []OHLC, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("timestamp,open,high,low,close,volume\n")
	if err != nil {
		return err
	}

	for _, ohlc := range data {
		_, err := file.WriteString(fmt.Sprintf("%d,%.5f,%.5f,%.5f,%.5f,%.5f\n",
			ohlc.Time, ohlc.Open, ohlc.High, ohlc.Low, ohlc.Close, ohlc.Volume))
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	pair := "XBTUSD"
	interval := 1440 // Daily
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	data, err := fetchHistoricalData(pair, interval, start, end)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}

	err = saveToCSV(data, "XBTUSD_daily_2023.csv")
	if err != nil {
		fmt.Println("Error saving to CSV:", err)
		return
	}

	fmt.Println("Data saved successfully")
}
