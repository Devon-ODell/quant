package strategy

import (
	"crypto-strats/internal/exchange"
	"fmt"
	"math"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Config holds configuration for the trading strategy
type Config struct {
	Symbol            string  `json:"symbol"`
	Period            int     `json:"period"`              // Bollinger Bands period
	StandardDev       float64 `json:"standard_dev"`        // Standard deviation multiplier
	BufferPercent     float64 `json:"buffer_percent"`      // Buffer percentage before bands
	EMAPeriod         int     `json:"ema_period"`          // Fast EMA period
	EMAPeriodSlow     int     `json:"ema_period_slow"`     // Slow EMA period
	MinTradeAmount    float64 `json:"min_trade_amount"`    // Minimum trade amount
	MaxPositionSize   float64 `json:"max_position_size"`   // Maximum position size
	StopLossPercent   float64 `json:"stop_loss_percent"`   // Stop loss percentage
	TakeProfitPercent float64 `json:"take_profit_percent"` // Take profit percentage
}

// Signal represents a trading signal
type Signal struct {
	Type       string          `json:"type"`       // BUY, SELL, HOLD
	Price      decimal.Decimal `json:"price"`      // Signal price
	Volume     decimal.Decimal `json:"volume"`     // Suggested volume
	Reason     string          `json:"reason"`     // Reason for the signal
	Timestamp  time.Time       `json:"timestamp"`  // Signal timestamp
	Confidence float64         `json:"confidence"` // Signal confidence (0-1)
}

// BollingerBandsEMA implements Bollinger Bands strategy with EMA crossover
type BollingerBandsEMA struct {
	config     *Config
	prices     []decimal.Decimal
	ema        []decimal.Decimal
	emaSlow    []decimal.Decimal
	sma        []decimal.Decimal
	upperBand  []decimal.Decimal
	lowerBand  []decimal.Decimal
	stdDev     []decimal.Decimal
	lastSignal *Signal
	position   *Position
}

// Position represents a current trading position
type Position struct {
	Symbol     string          `json:"symbol"`
	Type       string          `json:"type"` // LONG, SHORT
	EntryPrice decimal.Decimal `json:"entry_price"`
	Volume     decimal.Decimal `json:"volume"`
	EntryTime  time.Time       `json:"entry_time"`
	StopLoss   decimal.Decimal `json:"stop_loss"`
	TakeProfit decimal.Decimal `json:"take_profit"`
}

// NewBollingerBandsEMA creates a new Bollinger Bands + EMA strategy
func NewBollingerBandsEMA(config *Config) *BollingerBandsEMA {
	return &BollingerBandsEMA{
		config:    config,
		prices:    make([]decimal.Decimal, 0),
		ema:       make([]decimal.Decimal, 0),
		emaSlow:   make([]decimal.Decimal, 0),
		sma:       make([]decimal.Decimal, 0),
		upperBand: make([]decimal.Decimal, 0),
		lowerBand: make([]decimal.Decimal, 0),
		stdDev:    make([]decimal.Decimal, 0),
	}
}

// Update updates the strategy with new OHLC data
func (bb *BollingerBandsEMA) Update(ohlcData []exchange.OHLC) error {
	for _, ohlc := range ohlcData {
		bb.addPrice(ohlc.Close)
	}

	bb.calculateIndicators()
	return nil
}

// addPrice adds a new price to the price series
func (bb *BollingerBandsEMA) addPrice(price decimal.Decimal) {
	bb.prices = append(bb.prices, price)

	// Keep only the last 200 prices for efficiency
	if len(bb.prices) > 200 {
		bb.prices = bb.prices[1:]
	}
}

// calculateIndicators calculates all technical indicators
func (bb *BollingerBandsEMA) calculateIndicators() {
	if len(bb.prices) < bb.config.Period {
		return
	}

	// Calculate Simple Moving Average
	bb.calculateSMA()

	// Calculate Standard Deviation
	bb.calculateStandardDeviation()

	// Calculate Bollinger Bands
	bb.calculateBollingerBands()

	// Calculate EMAs
	bb.calculateEMA()
}

// calculateSMA calculates Simple Moving Average
func (bb *BollingerBandsEMA) calculateSMA() {
	if len(bb.prices) < bb.config.Period {
		return
	}

	start := len(bb.prices) - bb.config.Period
	sum := decimal.Zero

	for i := start; i < len(bb.prices); i++ {
		sum = sum.Add(bb.prices[i])
	}

	sma := sum.Div(decimal.NewFromInt(int64(bb.config.Period)))
	bb.sma = append(bb.sma, sma)

	// Keep only the last 100 SMA values
	if len(bb.sma) > 100 {
		bb.sma = bb.sma[1:]
	}
}

// calculateStandardDeviation calculates standard deviation for Bollinger Bands
func (bb *BollingerBandsEMA) calculateStandardDeviation() {
	if len(bb.prices) < bb.config.Period || len(bb.sma) == 0 {
		return
	}

	currentSMA := bb.sma[len(bb.sma)-1]
	start := len(bb.prices) - bb.config.Period

	sumSquaredDiff := decimal.Zero
	for i := start; i < len(bb.prices); i++ {
		diff := bb.prices[i].Sub(currentSMA)
		squaredDiff := diff.Mul(diff)
		sumSquaredDiff = sumSquaredDiff.Add(squaredDiff)
	}

	variance := sumSquaredDiff.Div(decimal.NewFromInt(int64(bb.config.Period)))
	stdDev, _ := decimal.NewFromString(fmt.Sprintf("%.8f", math.Sqrt(variance.InexactFloat64())))

	bb.stdDev = append(bb.stdDev, stdDev)

	// Keep only the last 100 standard deviation values
	if len(bb.stdDev) > 100 {
		bb.stdDev = bb.stdDev[1:]
	}
}

// calculateBollingerBands calculates Bollinger Bands
func (bb *BollingerBandsEMA) calculateBollingerBands() {
	if len(bb.sma) == 0 || len(bb.stdDev) == 0 {
		return
	}

	currentSMA := bb.sma[len(bb.sma)-1]
	currentStdDev := bb.stdDev[len(bb.stdDev)-1]

	stdDevMultiplier := decimal.NewFromFloat(bb.config.StandardDev)
	bandWidth := currentStdDev.Mul(stdDevMultiplier)

	upperBand := currentSMA.Add(bandWidth)
	lowerBand := currentSMA.Sub(bandWidth)

	bb.upperBand = append(bb.upperBand, upperBand)
	bb.lowerBand = append(bb.lowerBand, lowerBand)

	// Keep only the last 100 band values
	if len(bb.upperBand) > 100 {
		bb.upperBand = bb.upperBand[1:]
		bb.lowerBand = bb.lowerBand[1:]
	}
}

// calculateEMA calculates Exponential Moving Averages
func (bb *BollingerBandsEMA) calculateEMA() {
	if len(bb.prices) == 0 {
		return
	}

	currentPrice := bb.prices[len(bb.prices)-1]

	// Calculate fast EMA
	if len(bb.ema) == 0 {
		bb.ema = append(bb.ema, currentPrice)
	} else {
		alpha := decimal.NewFromFloat(2.0 / float64(bb.config.EMAPeriod+1))
		prevEMA := bb.ema[len(bb.ema)-1]
		newEMA := currentPrice.Mul(alpha).Add(prevEMA.Mul(decimal.NewFromInt(1).Sub(alpha)))
		bb.ema = append(bb.ema, newEMA)
	}

	// Calculate slow EMA
	if len(bb.emaSlow) == 0 {
		bb.emaSlow = append(bb.emaSlow, currentPrice)
	} else {
		alpha := decimal.NewFromFloat(2.0 / float64(bb.config.EMAPeriodSlow+1))
		prevEMA := bb.emaSlow[len(bb.emaSlow)-1]
		newEMA := currentPrice.Mul(alpha).Add(prevEMA.Mul(decimal.NewFromInt(1).Sub(alpha)))
		bb.emaSlow = append(bb.emaSlow, newEMA)
	}

	// Keep only the last 100 EMA values
	if len(bb.ema) > 100 {
		bb.ema = bb.ema[1:]
		bb.emaSlow = bb.emaSlow[1:]
	}
}

// GenerateSignal generates trading signals based on the strategy
func (bb *BollingerBandsEMA) GenerateSignal() *Signal {
	if len(bb.prices) < bb.config.Period ||
		len(bb.upperBand) == 0 ||
		len(bb.lowerBand) == 0 ||
		len(bb.ema) < 2 ||
		len(bb.emaSlow) < 2 {
		return &Signal{Type: "HOLD", Reason: "Insufficient data", Timestamp: time.Now()}
	}

	currentPrice := bb.prices[len(bb.prices)-1]
	upperBand := bb.upperBand[len(bb.upperBand)-1]
	lowerBand := bb.lowerBand[len(bb.lowerBand)-1]
	currentEMA := bb.ema[len(bb.ema)-1]
	prevEMA := bb.ema[len(bb.ema)-2]
	currentEMASlow := bb.emaSlow[len(bb.emaSlow)-1]
	prevEMASlow := bb.emaSlow[len(bb.emaSlow)-2]

	// Calculate 4% buffer zones
	bufferMultiplier := decimal.NewFromFloat(bb.config.BufferPercent / 100.0)
	upperBuffer := upperBand.Sub(upperBand.Sub(bb.sma[len(bb.sma)-1]).Mul(bufferMultiplier))
	lowerBuffer := lowerBand.Add(bb.sma[len(bb.sma)-1].Sub(lowerBand).Mul(bufferMultiplier))

	// EMA crossover signals
	emaRising := currentEMA.GreaterThan(prevEMA)
	emaSlowRising := currentEMASlow.GreaterThan(prevEMASlow)
	emaCrossoverBullish := currentEMA.GreaterThan(currentEMASlow) && prevEMA.LessThanOrEqual(prevEMASlow)
	emaCrossoverBearish := currentEMA.LessThan(currentEMASlow) && prevEMA.GreaterThanOrEqual(prevEMASlow)

	// Check for exit signals first if we have a position
	if bb.position != nil {
		if bb.position.Type == "LONG" {
			// Exit long position
			if currentPrice.LessThanOrEqual(bb.position.StopLoss) {
				return bb.createSignal("SELL", currentPrice, bb.position.Volume, "Stop loss triggered", 0.9)
			}
			if currentPrice.GreaterThanOrEqual(bb.position.TakeProfit) {
				return bb.createSignal("SELL", currentPrice, bb.position.Volume, "Take profit triggered", 0.9)
			}
			if emaCrossoverBearish || currentPrice.GreaterThan(upperBuffer) {
				return bb.createSignal("SELL", currentPrice, bb.position.Volume, "Exit signal: EMA bearish or upper band breach", 0.7)
			}
		} else if bb.position.Type == "SHORT" {
			// Exit short position
			if currentPrice.GreaterThanOrEqual(bb.position.StopLoss) {
				return bb.createSignal("BUY", currentPrice, bb.position.Volume, "Stop loss triggered", 0.9)
			}
			if currentPrice.LessThanOrEqual(bb.position.TakeProfit) {
				return bb.createSignal("BUY", currentPrice, bb.position.Volume, "Take profit triggered", 0.9)
			}
			if emaCrossoverBullish || currentPrice.LessThan(lowerBuffer) {
				return bb.createSignal("BUY", currentPrice, bb.position.Volume, "Exit signal: EMA bullish or lower band breach", 0.7)
			}
		}
	}

	// Entry signals
	confidence := bb.calculateConfidence(currentPrice, upperBand, lowerBand, emaRising, emaSlowRising)

	// Buy signal: Price approaches lower band with buffer + bullish EMA
	if currentPrice.LessThan(lowerBuffer) && emaRising && currentEMA.GreaterThan(currentEMASlow) {
		volume := bb.calculateVolume(currentPrice, "BUY")
		signal := bb.createSignal("BUY", currentPrice, volume, "Bollinger lower band approach + bullish EMA", confidence)
		bb.createPosition("LONG", currentPrice, volume)
		return signal
	}

	// Sell signal: Price approaches upper band with buffer + bearish EMA
	if currentPrice.GreaterThan(upperBuffer) && !emaRising && currentEMA.LessThan(currentEMASlow) {
		volume := bb.calculateVolume(currentPrice, "SELL")
		signal := bb.createSignal("SELL", currentPrice, volume, "Bollinger upper band approach + bearish EMA", confidence)
		bb.createPosition("SHORT", currentPrice, volume)
		return signal
	}

	// Strong EMA crossover signals
	if emaCrossoverBullish && currentPrice.GreaterThan(lowerBand) && currentPrice.LessThan(upperBand) {
		volume := bb.calculateVolume(currentPrice, "BUY")
		return bb.createSignal("BUY", currentPrice, volume, "EMA bullish crossover within bands", confidence*0.8)
	}

	if emaCrossoverBearish && currentPrice.GreaterThan(lowerBand) && currentPrice.LessThan(upperBand) {
		volume := bb.calculateVolume(currentPrice, "SELL")
		return bb.createSignal("SELL", currentPrice, volume, "EMA bearish crossover within bands", confidence*0.8)
	}

	return &Signal{Type: "HOLD", Reason: "No strong signals", Timestamp: time.Now(), Confidence: confidence}
}

// createSignal creates a new trading signal
func (bb *BollingerBandsEMA) createSignal(signalType string, price, volume decimal.Decimal, reason string, confidence float64) *Signal {
	signal := &Signal{
		Type:       signalType,
		Price:      price,
		Volume:     volume,
		Reason:     reason,
		Timestamp:  time.Now(),
		Confidence: confidence,
	}

	bb.lastSignal = signal

	logrus.WithFields(logrus.Fields{
		"symbol":     bb.config.Symbol,
		"type":       signalType,
		"price":      price,
		"volume":     volume,
		"reason":     reason,
		"confidence": confidence,
	}).Info("Generated trading signal")

	return signal
}

// createPosition creates a new position
func (bb *BollingerBandsEMA) createPosition(posType string, entryPrice, volume decimal.Decimal) {
	stopLossPercent := decimal.NewFromFloat(bb.config.StopLossPercent / 100.0)
	takeProfitPercent := decimal.NewFromFloat(bb.config.TakeProfitPercent / 100.0)

	var stopLoss, takeProfit decimal.Decimal

	if posType == "LONG" {
		stopLoss = entryPrice.Mul(decimal.NewFromInt(1).Sub(stopLossPercent))
		takeProfit = entryPrice.Mul(decimal.NewFromInt(1).Add(takeProfitPercent))
	} else {
		stopLoss = entryPrice.Mul(decimal.NewFromInt(1).Add(stopLossPercent))
		takeProfit = entryPrice.Mul(decimal.NewFromInt(1).Sub(takeProfitPercent))
	}

	bb.position = &Position{
		Symbol:     bb.config.Symbol,
		Type:       posType,
		EntryPrice: entryPrice,
		Volume:     volume,
		EntryTime:  time.Now(),
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	}
}

// calculateVolume calculates the volume for a trade based on risk management
func (bb *BollingerBandsEMA) calculateVolume(price decimal.Decimal, orderType string) decimal.Decimal {
	maxAmount := decimal.NewFromFloat(bb.config.MaxPositionSize)
	minAmount := decimal.NewFromFloat(bb.config.MinTradeAmount)

	// Calculate volume based on price and available capital
	volume := maxAmount.Div(price)

	// Ensure minimum trade amount
	minVolume := minAmount.Div(price)
	if volume.LessThan(minVolume) {
		volume = minVolume
	}

	return volume
}

// calculateConfidence calculates the confidence level of the signal
func (bb *BollingerBandsEMA) calculateConfidence(price, upperBand, lowerBand decimal.Decimal, emaRising, emaSlowRising bool) float64 {
	confidence := 0.5 // Base confidence

	// Distance from bands increases confidence
	bandWidth := upperBand.Sub(lowerBand)
	middleBand := bb.sma[len(bb.sma)-1]

	distanceFromMiddle := price.Sub(middleBand).Abs()
	relativeDistance := distanceFromMiddle.Div(bandWidth).InexactFloat64()

	// More distance from center = higher confidence
	confidence += relativeDistance * 0.3

	// EMA alignment increases confidence
	if emaRising && emaSlowRising {
		confidence += 0.2
	} else if !emaRising && !emaSlowRising {
		confidence += 0.2
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// GetPosition returns the current position
func (bb *BollingerBandsEMA) GetPosition() *Position {
	return bb.position
}

// ClosePosition closes the current position
func (bb *BollingerBandsEMA) ClosePosition() {
	if bb.position != nil {
		logrus.WithFields(logrus.Fields{
			"symbol":     bb.position.Symbol,
			"type":       bb.position.Type,
			"entryPrice": bb.position.EntryPrice,
			"volume":     bb.position.Volume,
		}).Info("Position closed")

		bb.position = nil
	}
}

// GetConfig returns the strategy configuration
func (bb *BollingerBandsEMA) GetConfig() *Config {
	return bb.config
}
