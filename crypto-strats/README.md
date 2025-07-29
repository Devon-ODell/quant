# Crypto Trading Strategy Bot

A sophisticated Go-based cryptocurrency trading bot that implements Bollinger Bands strategy with 4% buffer zones and EMA (Exponential Moving Average) crossover signals for trading LINK/USD and ETH/USD on Kraken exchange.

## Features

### Trading Strategy
- **Bollinger Bands**: 1.5 standard deviation bands with 4% buffer zones before entry/exit
- **EMA Crossover**: Dual EMA system (12-period fast, 26-period slow) for trend confirmation
- **Agentic Analysis**: Intelligent signal processing with confidence scoring
- **Risk Management**: Built-in stop-loss and take-profit levels
- **Position Management**: Automatic position tracking and exit signals

### Technical Indicators
- **Bollinger Bands**: Moving average with standard deviation bands
- **Exponential Moving Average (EMA)**: Fast and slow EMA for trend analysis
- **Buffer Zones**: 4% buffer before band perimeters to avoid false signals
- **Confidence Scoring**: Signal strength evaluation based on multiple factors

### Supported Assets
- **Chainlink (LINK/USD)**
- **Ethereum (ETH/USD)**

## Strategy Logic

### Entry Signals
1. **Long Entry**: Price approaches lower Bollinger Band (with 4% buffer) + bullish EMA crossover
2. **Short Entry**: Price approaches upper Bollinger Band (with 4% buffer) + bearish EMA crossover
3. **EMA Crossover**: Strong fast EMA crossing above/below slow EMA within bands

### Exit Signals
1. **Stop Loss**: Configurable percentage-based stop loss
2. **Take Profit**: Configurable percentage-based take profit
3. **Band Breach**: Price breaking through buffer zones
4. **EMA Reversal**: EMA crossover in opposite direction

### Risk Management
- Maximum position size limits
- Minimum trade amount enforcement
- Portfolio-wide risk limits
- Daily loss limits
- Position count limits

## Setup

### Prerequisites
- Go 1.21 or higher
- Kraken exchange account with API access
- API key and private key from Kraken

### Installation

1. **Clone or create the project directory:**
```bash
cd "c:\Users\Devon Odell\Desktop\quant\crypto-strats"
```

2. **Initialize Go module and install dependencies:**
```bash
go mod init crypto-strats
go mod tidy
```

3. **Set up environment variables:**

Create a `.env` file or set environment variables:
```bash
# Windows PowerShell
$env:KRAKEN_API_KEY="your_kraken_api_key"
$env:KRAKEN_PRIVATE_KEY="your_kraken_private_key"

# Or set them permanently in Windows
setx KRAKEN_API_KEY "your_kraken_api_key"
setx KRAKEN_PRIVATE_KEY "your_kraken_private_key"
```

4. **Configure trading parameters:**

Edit `config.json` to customize risk management settings:
```json
{
  "kraken": {
    "api_key": "",
    "private_key": "",
    "base_url": "https://api.kraken.com"
  },
  "risk": {
    "max_portfolio_risk": 0.02,
    "max_daily_loss": 0.05,
    "max_positions": 4
  }
}
```

### Kraken API Setup

1. **Log into your Kraken account**
2. **Go to Settings > API**
3. **Create a new API key with permissions:**
   - Query Funds
   - Query Open Orders
   - Query Closed Orders
   - Query Trades History
   - Create & Modify Orders
   - Cancel Orders

4. **Copy the API Key and Private Key**
5. **Set them as environment variables**

## Usage

### Running the Bot

```bash
# Build the application
go build -o crypto-trader.exe

# Run the trading bot
./crypto-trader.exe
```

### Running in Development Mode

```bash
# Run directly with Go
go run main.go
```

The bot will:
1. Initialize with historical data (24 hours)
2. Calculate Bollinger Bands and EMA indicators
3. Monitor price movements every 30 seconds
4. Generate trading signals based on the strategy
5. Execute trades when confidence > 70%
6. Log all activities and decisions

### Monitoring

The bot provides detailed logging including:
- Signal generation with reasons and confidence scores
- Trade execution confirmations
- Position updates and risk management actions
- Error handling and API issues

## Strategy Configuration

### LINK/USD Configuration
```go
linkStrategy := strategy.NewBollingerBandsEMA(&strategy.Config{
    Symbol:           "LINKUSD",
    Period:           20,      // Bollinger Bands period
    StandardDev:      1.5,     // Standard deviation multiplier
    BufferPercent:    4.0,     // 4% buffer before bands
    EMAPeriod:        12,      // Fast EMA
    EMAPeriodSlow:    26,      // Slow EMA
    MinTradeAmount:   10.0,    // Minimum trade size
    MaxPositionSize:  1000.0,  // Maximum position size
    StopLossPercent:  5.0,     // 5% stop loss
    TakeProfitPercent: 8.0,    // 8% take profit
})
```

### ETH/USD Configuration
```go
ethStrategy := strategy.NewBollingerBandsEMA(&strategy.Config{
    Symbol:           "ETHUSD",
    Period:           20,
    StandardDev:      1.5,
    BufferPercent:    4.0,
    EMAPeriod:        12,
    EMAPeriodSlow:    26,
    MinTradeAmount:   50.0,    // Higher minimum for ETH
    MaxPositionSize:  5000.0,  // Higher maximum for ETH
    StopLossPercent:  4.0,     // Tighter stop loss
    TakeProfitPercent: 7.0,    // Adjusted take profit
})
```

## Project Structure

```
crypto-strats/
├── main.go                           # Application entry point
├── go.mod                           # Go module definition
├── config.json                      # Configuration file
├── internal/
│   ├── agent/
│   │   └── trading_agent.go         # Main trading orchestration
│   ├── config/
│   │   └── config.go                # Configuration management
│   ├── exchange/
│   │   └── kraken.go                # Kraken API client
│   └── strategy/
│       └── bollinger_ema.go         # Bollinger Bands + EMA strategy
└── README.md                        # This file
```

## Safety Features

### Risk Management
- **Position Limits**: Maximum position sizes per asset
- **Portfolio Risk**: Overall portfolio risk limits
- **Daily Loss Limits**: Stop trading if daily losses exceed threshold
- **Confidence Thresholds**: Only execute high-confidence signals

### Error Handling
- **API Failures**: Graceful handling of exchange API issues
- **Network Issues**: Retry logic and timeout handling
- **Invalid Data**: Data validation and sanitization
- **Order Failures**: Proper error reporting and recovery

### Logging
- **Structured Logging**: JSON-formatted logs with contextual information
- **Trade History**: Complete audit trail of all trading decisions
- **Performance Metrics**: Strategy performance tracking
- **Error Tracking**: Detailed error logging and reporting

## Disclaimer

⚠️ **IMPORTANT TRADING DISCLAIMER** ⚠️

This trading bot is for educational and research purposes. Cryptocurrency trading involves significant financial risk:

- **Past performance does not guarantee future results**
- **You can lose all or more than your initial investment**
- **Start with small amounts and paper trading**
- **Understand the strategy before risking real money**
- **Monitor the bot's performance continuously**
- **Be prepared to stop the bot if losses mount**

### Recommended Testing Approach

1. **Paper Trading**: Test with simulated funds first
2. **Small Amounts**: Start with minimal real capital
3. **Monitor Closely**: Watch the first few days/weeks carefully
4. **Backtest**: Test the strategy on historical data
5. **Gradual Scaling**: Only increase capital after proven performance

## Contributing

Feel free to improve the strategy, add new features, or fix bugs. Some areas for enhancement:

- Additional technical indicators
- More sophisticated risk management
- Backtesting framework
- Web dashboard for monitoring
- Support for more trading pairs
- Machine learning signal enhancement

## License

This project is for educational purposes. Use at your own risk.
