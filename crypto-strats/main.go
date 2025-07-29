package main

import (
	"crypto-strats/internal/agent"
	"crypto-strats/internal/config"
	"crypto-strats/internal/exchange"
	"crypto-strats/internal/strategy"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kraken exchange client
	kraken := exchange.NewKrakenClient(cfg.Kraken.APIKey, cfg.Kraken.PrivateKey, cfg.Kraken.BaseURL)

	// Initialize trading strategies
	linkStrategy := strategy.NewBollingerBandsEMA(&strategy.Config{
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
	})

	ethStrategy := strategy.NewBollingerBandsEMA(&strategy.Config{
		Symbol:            "ETHUSD",
		Period:            20,
		StandardDev:       1.5,
		BufferPercent:     4.0,
		EMAPeriod:         12,
		EMAPeriodSlow:     26,
		MinTradeAmount:    50.0,
		MaxPositionSize:   5000.0,
		StopLossPercent:   4.0,
		TakeProfitPercent: 7.0,
	})

	// Initialize trading agent
	tradingAgent := agent.NewTradingAgent(kraken, []*strategy.BollingerBandsEMA{
		linkStrategy,
		ethStrategy,
	})

	// Start the trading agent
	go tradingAgent.Start()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logrus.Info("Crypto trading bot started. Trading LINK/USD and ETH/USD...")
	logrus.Info("Strategies: Bollinger Bands (1.5σ) with 4% buffer + EMA crossover")

	<-sigChan
	logrus.Info("Shutting down trading bot...")
	tradingAgent.Stop()
	time.Sleep(2 * time.Second)
	logrus.Info("Trading bot stopped")
}
