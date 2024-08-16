package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type OHLC struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type Strategy interface {
	Initialize(data []OHLC)
	OnData(current OHLC, position *Position) (action string, amount float64)
}

type Position struct {
	Asset  string
	Amount float64
}

type SimpleMovingAverageStrategy struct {
	Data     []OHLC
	MAPeriod int
	MA       []float64
}

func (s *SimpleMovingAverageStrategy) Initialize(data []OHLC) {
	s.Data = data
	s.MA = make([]float64, len(data))

	for i := s.MAPeriod - 1; i < len(data); i++ {
		sum := 0.0
		for j := 0; j < s.MAPeriod; j++ {
			sum += s.Data[i-j].Close
		}
		s.MA[i] = sum / float64(s.MAPeriod)
	}
}

func (s *SimpleMovingAverageStrategy) OnData(current OHLC, position *Position) (string, float64) {
	index := len(s.Data) - 1
	if current.Close > s.MA[index] && position.Amount == 0 {
		return "buy", 1.0
	} else if current.Close < s.MA[index] && position.Amount > 0 {
		return "sell", position.Amount
	}
	return "hold", 0
}

type Backtest struct {
	Data     []OHLC
	Strategy Strategy
	Cash     float64
	Position Position
}

func NewBacktest(data []OHLC, strategy Strategy, initialCash float64) *Backtest {
	return &Backtest{
		Data:     data,
		Strategy: strategy,
		Cash:     initialCash,
		Position: Position{Asset: "BTC", Amount: 0},
	}
}

func (b *Backtest) Run() {
	b.Strategy.Initialize(b.Data)

	for _, ohlc := range b.Data {
		action, amount := b.Strategy.OnData(ohlc, &b.Position)

		switch action {
		case "buy":
			if b.Cash > 0 {
				buyAmount := amount * ohlc.Close
				if buyAmount > b.Cash {
					buyAmount = b.Cash
				}
				b.Position.Amount += buyAmount / ohlc.Close
				b.Cash -= buyAmount
			}
		case "sell":
			if b.Position.Amount > 0 {
				sellAmount := amount * ohlc.Close
				b.Cash += sellAmount
				b.Position.Amount -= amount
			}
		}
	}
}

func (b *Backtest) Results() {
	initialValue := b.Cash
	finalValue := b.Cash + b.Position.Amount*b.Data[len(b.Data)-1].Close
	profit := finalValue - initialValue
	returnPercentage := (finalValue / initialValue - 1) * 100

	fmt.Printf("Initial Portfolio Value: $%.2f\n", initialValue)
	fmt.Printf("Final Portfolio Value: $%.2f\n", finalValue)
	fmt.Printf("Profit/Loss: $%.2f\n", profit)
	fmt.Printf("Return: %.2f%%\n", returnPercentage)
}

func loadCSV(filename string) ([]OHLC, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var data []OHLC
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}

		timestamp, _ := strconv.ParseInt(record[0], 10, 64)
		open, _ := strconv.ParseFloat(record[1], 64)
		high, _ := strconv.ParseFloat(record[2], 64)
		low, _ := strconv.ParseFloat(record[3], 64)
		close, _ := strconv.ParseFloat(record[4], 64)
		volume, _ := strconv.ParseFloat(record[5], 64)

		data = append(data, OHLC{
			Time:   time.Unix(timestamp, 0),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return data, nil
}

func main() {
	data, err := loadCSV("XBTUSD_daily_2023.csv")
	if err != nil {
		fmt.Println("Error loading CSV:", err)
		return
	}

	strategy := &SimpleMovingAverageStrategy{MAPeriod: 20}
	backtest := NewBacktest(data, strategy, 10000)
	backtest.Run()
	backtest.Results()
}
