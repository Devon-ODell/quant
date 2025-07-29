package agent

import (
	"crypto-strats/internal/exchange"
	"crypto-strats/internal/strategy"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TradingAgent represents the main trading agent
type TradingAgent struct {
	exchange   *exchange.KrakenClient
	strategies []*strategy.BollingerBandsEMA
	running    bool
	stopChan   chan bool
	mutex      sync.RWMutex
}

// NewTradingAgent creates a new trading agent
func NewTradingAgent(ex *exchange.KrakenClient, strats []*strategy.BollingerBandsEMA) *TradingAgent {
	return &TradingAgent{
		exchange:   ex,
		strategies: strats,
		stopChan:   make(chan bool),
	}
}

// Start starts the trading agent
func (ta *TradingAgent) Start() {
	ta.mutex.Lock()
	ta.running = true
	ta.mutex.Unlock()

	logrus.Info("Trading agent started")

	// Initialize strategies with historical data
	ta.initializeStrategies()

	// Start main trading loop
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ta.stopChan:
			logrus.Info("Trading agent stopped")
			return
		case <-ticker.C:
			ta.executeTradeLoop()
		}
	}
}

// Stop stops the trading agent
func (ta *TradingAgent) Stop() {
	ta.mutex.Lock()
	defer ta.mutex.Unlock()

	if ta.running {
		ta.running = false
		close(ta.stopChan)
	}
}

// executeTradeLoop executes one iteration of the trading loop
func (ta *TradingAgent) executeTradeLoop() {
	ta.mutex.RLock()
	running := ta.running
	ta.mutex.RUnlock()

	if !running {
		return
	}

	for _, strat := range ta.strategies {
		ta.processStrategy(strat)
	}
}

// processStrategy processes a single strategy
func (ta *TradingAgent) processStrategy(strat *strategy.BollingerBandsEMA) {
	config := strat.GetConfig()

	// Get latest OHLC data
	ohlcData, err := ta.exchange.GetOHLC(config.Symbol, 1, time.Now().Add(-24*time.Hour))
	if err != nil {
		logrus.WithError(err).WithField("symbol", config.Symbol).Error("Failed to get OHLC data")
		return
	}

	if len(ohlcData) == 0 {
		logrus.WithField("symbol", config.Symbol).Warn("No OHLC data received")
		return
	}

	// Update strategy with new data
	err = strat.Update(ohlcData)
	if err != nil {
		logrus.WithError(err).WithField("symbol", config.Symbol).Error("Failed to update strategy")
		return
	}

	// Generate signal
	signal := strat.GenerateSignal()
	if signal == nil {
		return
	}

	// Process signal
	ta.processSignal(strat, signal)
}

// processSignal processes a trading signal
func (ta *TradingAgent) processSignal(strat *strategy.BollingerBandsEMA, signal *strategy.Signal) {
	config := strat.GetConfig()

	logrus.WithFields(logrus.Fields{
		"symbol":     config.Symbol,
		"signal":     signal.Type,
		"price":      signal.Price,
		"volume":     signal.Volume,
		"reason":     signal.Reason,
		"confidence": signal.Confidence,
	}).Info("Processing signal")

	// Only execute trades with high confidence
	if signal.Confidence < 0.7 {
		logrus.WithField("confidence", signal.Confidence).Info("Signal confidence too low, skipping trade")
		return
	}

	if signal.Type == "BUY" || signal.Type == "SELL" {
		err := ta.executeOrder(strat, signal)
		if err != nil {
			logrus.WithError(err).WithField("symbol", config.Symbol).Error("Failed to execute order")
		}
	}
}

// executeOrder executes a trading order
func (ta *TradingAgent) executeOrder(strat *strategy.BollingerBandsEMA, signal *strategy.Signal) error {
	config := strat.GetConfig()

	// Create order request
	orderReq := &exchange.OrderRequest{
		Pair:      config.Symbol,
		Type:      signal.Type,
		OrderType: "market",
		Volume:    signal.Volume,
	}

	// For limit orders, you could set a price slightly better than market
	// orderReq.OrderType = "limit"
	// orderReq.Price = signal.Price

	// Execute the order
	response, err := ta.exchange.PlaceOrder(orderReq)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"symbol":   config.Symbol,
		"type":     signal.Type,
		"volume":   signal.Volume,
		"response": response.Result,
	}).Info("Order executed successfully")

	// Update position tracking
	if signal.Type == "SELL" {
		strat.ClosePosition()
	}

	return nil
}

// initializeStrategies initializes all strategies with historical data
func (ta *TradingAgent) initializeStrategies() {
	logrus.Info("Initializing strategies with historical data...")

	for _, strat := range ta.strategies {
		config := strat.GetConfig()

		// Get 24 hours of historical data for initialization
		ohlcData, err := ta.exchange.GetOHLC(config.Symbol, 1, time.Now().Add(-24*time.Hour))
		if err != nil {
			logrus.WithError(err).WithField("symbol", config.Symbol).Error("Failed to get historical data")
			continue
		}

		err = strat.Update(ohlcData)
		if err != nil {
			logrus.WithError(err).WithField("symbol", config.Symbol).Error("Failed to initialize strategy")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"symbol":     config.Symbol,
			"dataPoints": len(ohlcData),
		}).Info("Strategy initialized")
	}
}

// GetStrategies returns all strategies
func (ta *TradingAgent) GetStrategies() []*strategy.BollingerBandsEMA {
	return ta.strategies
}

// IsRunning returns whether the agent is running
func (ta *TradingAgent) IsRunning() bool {
	ta.mutex.RLock()
	defer ta.mutex.RUnlock()
	return ta.running
}
