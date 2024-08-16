package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"timestamp", "open", "high", "low", "close", "volume"}); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Write data
	for _, ohlc := range data {
		record := []string{
			strconv.FormatInt(ohlc.Time, 10),
			strconv.FormatFloat(ohlc.Open, 'f', 5, 64),
			strconv.FormatFloat(ohlc.High, 'f', 5, 64),
			strconv.FormatFloat(ohlc.Low, 'f', 5, 64),
			strconv.FormatFloat(ohlc.Close, 'f', 5, 64),
			strconv.FormatFloat(ohlc.Volume, 'f', 5, 64),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
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
		fmt.Printf("Error fetching data: %v\n", err)
		return
	}

	filename := filepath.Join("..", "data", fmt.Sprintf("%s_daily_%d.csv", pair, start.Year()))
	err = saveToCSV(data, filename)
	if err != nil {
		fmt.Printf("Error saving to CSV: %v\n", err)
		return
	}

	fmt.Printf("Data saved successfully to %s\n", filename)
}
