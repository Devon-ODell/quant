// +build ignore

package main

import (
	"crypto-strats/internal/exchange"
	"crypto-strats/internal/strategy"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// Example program to test the trading strategy without live trading
// To run this test:
// 1. Rename main.go to main_live.go
// 2. Rename this file to main.go
// 3. Run: go run main.go
func main() {
	fmt.Println("=== Crypto Trading Strategy Test ===")
	fmt.Println()

	// Create LINK strategy configuration
	linkConfig := &strategy.Config{
		Symbol:            "LINKUSD",
		Period:            20,
		StandardDev:       1.5,
		BufferPercent:     4.0,
		EMAPeriod:         12,
		EMAPeriodSlow:     26,
		MinTradeAmount:    10.0,
		MaxPositionSize:   1000.0,
		StopLossPercent:   5.0,
		TakeProfitPercent: 8.0,
	}

	// Create strategy instance
	linkStrategy := strategy.NewBollingerBandsEMA(linkConfig)

	// Simulate some price data for testing
	testPrices := []decimal.Decimal{
		decimal.NewFromFloat(15.50),
		decimal.NewFromFloat(15.75),
		decimal.NewFromFloat(15.60),
		decimal.NewFromFloat(15.80),
		decimal.NewFromFloat(15.95),
		decimal.NewFromFloat(16.10),
		decimal.NewFromFloat(16.25),
		decimal.NewFromFloat(16.40),
		decimal.NewFromFloat(16.20),
		decimal.NewFromFloat(16.05),
		decimal.NewFromFloat(15.90),
		decimal.NewFromFloat(15.75),
		decimal.NewFromFloat(15.60),
		decimal.NewFromFloat(15.45),
		decimal.NewFromFloat(15.30),
		decimal.NewFromFloat(15.40),
		decimal.NewFromFloat(15.55),
		decimal.NewFromFloat(15.70),
		decimal.NewFromFloat(15.85),
		decimal.NewFromFloat(16.00),
		decimal.NewFromFloat(16.15),
		decimal.NewFromFloat(16.30),
		decimal.NewFromFloat(16.45),
		decimal.NewFromFloat(16.60),
		decimal.NewFromFloat(16.75),
	}

	// Create mock OHLC data
	var ohlcData []exchange.OHLC
	baseTime := time.Now().Add(-25 * time.Hour)

	for i, price := range testPrices {
		ohlc := exchange.OHLC{
			Time:   baseTime.Add(time.Duration(i) * time.Hour),
			Open:   price.Sub(decimal.NewFromFloat(0.05)),
			High:   price.Add(decimal.NewFromFloat(0.10)),
			Low:    price.Sub(decimal.NewFromFloat(0.10)),
			Close:  price,
			Volume: decimal.NewFromFloat(1000.0),
		}
		ohlcData = append(ohlcData, ohlc)
	}

	fmt.Printf("Testing with %d price points...\n", len(testPrices))
	fmt.Println()

	// Update strategy with test data
	err := linkStrategy.Update(ohlcData)
	if err != nil {
		log.Fatalf("Failed to update strategy: %v", err)
	}

	// Generate and display signals
	signal := linkStrategy.GenerateSignal()
	if signal != nil {
		fmt.Printf("Generated Signal:\n")
		fmt.Printf("  Type: %s\n", signal.Type)
		fmt.Printf("  Price: $%s\n", signal.Price.StringFixed(2))
		fmt.Printf("  Volume: %s\n", signal.Volume.StringFixed(4))
		fmt.Printf("  Reason: %s\n", signal.Reason)
		fmt.Printf("  Confidence: %.1f%%\n", signal.Confidence*100)
		fmt.Printf("  Timestamp: %s\n", signal.Timestamp.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("No signal generated")
	}

	fmt.Println()

	// Display current position if any
	position := linkStrategy.GetPosition()
	if position != nil {
		fmt.Printf("Current Position:\n")
		fmt.Printf("  Symbol: %s\n", position.Symbol)
		fmt.Printf("  Type: %s\n", position.Type)
		fmt.Printf("  Entry Price: $%s\n", position.EntryPrice.StringFixed(2))
		fmt.Printf("  Volume: %s\n", position.Volume.StringFixed(4))
		fmt.Printf("  Stop Loss: $%s\n", position.StopLoss.StringFixed(2))
		fmt.Printf("  Take Profit: $%s\n", position.TakeProfit.StringFixed(2))
		fmt.Printf("  Entry Time: %s\n", position.EntryTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("No current position")
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
	fmt.Println()
	fmt.Println("To run with live data:")
	fmt.Println("1. Rename this file and rename main_live.go to main.go")
	fmt.Println("2. Set your Kraken API credentials as environment variables")
	fmt.Println("3. Run: go run main.go")
	fmt.Println()
	fmt.Println("⚠️  Remember to start with small amounts and monitor carefully!")
}
