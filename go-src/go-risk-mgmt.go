package main

import (
	"math"
)

// CalculatePositionSize determines the position size based on account balance and risk percentage
func CalculatePositionSize(accountBalance, riskPercentage, entryPrice, stopLossPrice float64) float64 {
	riskAmount := accountBalance * riskPercentage
	riskPerShare := math.Abs(entryPrice - stopLossPrice)
	return riskAmount / riskPerShare
}

// PlaceStopLoss calculates the stop-loss price based on entry price and risk percentage
func PlaceStopLoss(entryPrice, riskPercentage float64) float64 {
	return entryPrice * (1 - riskPercentage)
}

// CheckRiskRewardRatio ensures the trade has at least a 1:2 risk-reward ratio
func CheckRiskRewardRatio(entryPrice, stopLossPrice, takeProfitPrice float64) bool {
	risk := math.Abs(entryPrice - stopLossPrice)
	reward := math.Abs(takeProfitPrice - entryPrice)
	ratio := reward / risk
	return ratio >= 2
}

// CheckMaxDrawdown verifies if the current drawdown exceeds the maximum allowed
func CheckMaxDrawdown(initialBalance, currentBalance, maxDrawdownPercentage float64) bool {
	drawdown := (initialBalance - currentBalance) / initialBalance
	return drawdown <= maxDrawdownPercentage
}

// Asset represents a tradable asset
type Asset struct {
	Symbol string
	Returns []float64
}

// CheckCorrelation ensures that assets in the portfolio are not highly correlated
func CheckCorrelation(assets []Asset, threshold float64) bool {
	for i := 0; i < len(assets); i++ {
		for j := i + 1; j < len(assets); j++ {
			if correlation(assets[i].Returns, assets[j].Returns) > threshold {
				return false
			}
		}
	}
	return true
}

// correlation calculates the Pearson correlation coefficient between two sets of returns
func correlation(x, y []float64) float64 {
	n := float64(len(x))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := sumXY - (sumX * sumY / n)
	denominator := math.Sqrt((sumX2 - (sumX * sumX / n)) * (sumY2 - (sumY * sumY / n)))

	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// Example usage
func main() {
	// Position sizing example
	accountBalance := 10000.0
	riskPercentage := 0.01 // 1%
	entryPrice := 100.0
	stopLossPrice := 98.0

	positionSize := CalculatePositionSize(accountBalance, riskPercentage, entryPrice, stopLossPrice)
	println("Position size:", positionSize)

	// Stop-loss example
	stopLoss := PlaceStopLoss(entryPrice, 0.02) // 2% risk
	println("Stop-loss price:", stopLoss)

	// Risk-reward ratio example
	takeProfitPrice := 105.0
	goodRiskReward := CheckRiskRewardRatio(entryPrice, stopLossPrice, takeProfitPrice)
	println("Good risk-reward ratio:", goodRiskReward)

	// Max drawdown example
	currentBalance := 9000.0
	maxDrawdownPercentage := 0.20 // 20%
	withinMaxDrawdown := CheckMaxDrawdown(accountBalance, currentBalance, maxDrawdownPercentage)
	println("Within max drawdown:", withinMaxDrawdown)

	// Correlation example
	assets := []Asset{
		{Symbol: "BTC", Returns: []float64{0.01, -0.02, 0.03, 0.01}},
		{Symbol: "ETH", Returns: []float64{0.02, -0.01, 0.02, 0.01}},
		{Symbol: "XRP", Returns: []float64{-0.01, 0.01, -0.02, -0.01}},
	}
	lowCorrelation := CheckCorrelation(assets, 0.7)
	println("Low correlation between assets:", lowCorrelation)
}
