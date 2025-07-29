package exchange

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// KrakenClient represents a client for Kraken exchange
type KrakenClient struct {
	apiKey     string
	privateKey string
	baseURL    string
	httpClient *http.Client
}

// OHLC represents OHLC data from Kraken
type OHLC struct {
	Time   time.Time       `json:"time"`
	Open   decimal.Decimal `json:"open"`
	High   decimal.Decimal `json:"high"`
	Low    decimal.Decimal `json:"low"`
	Close  decimal.Decimal `json:"close"`
	Volume decimal.Decimal `json:"volume"`
}

// OrderRequest represents an order request
type OrderRequest struct {
	Pair      string          `json:"pair"`
	Type      string          `json:"type"`      // buy/sell
	OrderType string          `json:"ordertype"` // market/limit
	Volume    decimal.Decimal `json:"volume"`
	Price     decimal.Decimal `json:"price,omitempty"`
}

// OrderResponse represents an order response from Kraken
type OrderResponse struct {
	Error  []string               `json:"error"`
	Result map[string]interface{} `json:"result"`
}

// TickerResponse represents ticker data from Kraken
type TickerResponse struct {
	Error  []string                          `json:"error"`
	Result map[string]map[string]interface{} `json:"result"`
}

// NewKrakenClient creates a new Kraken client
func NewKrakenClient(apiKey, privateKey, baseURL string) *KrakenClient {
	return &KrakenClient{
		apiKey:     apiKey,
		privateKey: privateKey,
		baseURL:    baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetOHLC fetches OHLC data for a trading pair
func (k *KrakenClient) GetOHLC(pair string, interval int, since time.Time) ([]OHLC, error) {
	endpoint := "/0/public/OHLC"
	params := url.Values{}
	params.Set("pair", pair)
	params.Set("interval", strconv.Itoa(interval))
	if !since.IsZero() {
		params.Set("since", strconv.FormatInt(since.Unix(), 10))
	}

	resp, err := k.makeRequest("GET", endpoint, params, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get OHLC data: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OHLC response: %w", err)
	}

	if errors, ok := response["error"].([]interface{}); ok && len(errors) > 0 {
		return nil, fmt.Errorf("kraken API error: %v", errors)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	// Find the pair data in the result
	var ohlcData []interface{}
	for key, value := range result {
		if key != "last" {
			if data, ok := value.([]interface{}); ok {
				ohlcData = data
				break
			}
		}
	}

	if ohlcData == nil {
		return nil, fmt.Errorf("no OHLC data found for pair %s", pair)
	}

	var ohlcs []OHLC
	for _, item := range ohlcData {
		if arr, ok := item.([]interface{}); ok && len(arr) >= 6 {
			timestamp, _ := strconv.ParseFloat(arr[0].(string), 64)
			open, _ := decimal.NewFromString(arr[1].(string))
			high, _ := decimal.NewFromString(arr[2].(string))
			low, _ := decimal.NewFromString(arr[3].(string))
			close, _ := decimal.NewFromString(arr[4].(string))
			volume, _ := decimal.NewFromString(arr[6].(string))

			ohlc := OHLC{
				Time:   time.Unix(int64(timestamp), 0),
				Open:   open,
				High:   high,
				Low:    low,
				Close:  close,
				Volume: volume,
			}
			ohlcs = append(ohlcs, ohlc)
		}
	}

	return ohlcs, nil
}

// GetTicker gets current ticker information
func (k *KrakenClient) GetTicker(pair string) (decimal.Decimal, error) {
	endpoint := "/0/public/Ticker"
	params := url.Values{}
	params.Set("pair", pair)

	resp, err := k.makeRequest("GET", endpoint, params, false)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get ticker: %w", err)
	}

	var response TickerResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return decimal.Zero, fmt.Errorf("failed to unmarshal ticker response: %w", err)
	}

	if len(response.Error) > 0 {
		return decimal.Zero, fmt.Errorf("kraken API error: %v", response.Error)
	}

	// Extract current price from the ticker data
	for _, tickerData := range response.Result {
		if closeData, ok := tickerData["c"]; ok {
			// closeData is an array where first element is the price
			if closeArray, ok := closeData.([]interface{}); ok && len(closeArray) > 0 {
				if priceStr, ok := closeArray[0].(string); ok {
					price, err := decimal.NewFromString(priceStr)
					if err != nil {
						return decimal.Zero, fmt.Errorf("failed to parse price: %w", err)
					}
					return price, nil
				}
			} else if priceStr, ok := closeData.(string); ok {
				price, err := decimal.NewFromString(priceStr)
				if err != nil {
					return decimal.Zero, fmt.Errorf("failed to parse price: %w", err)
				}
				return price, nil
			}
		}
	}

	return decimal.Zero, fmt.Errorf("no price data found for pair %s", pair)
}

// PlaceOrder places a trading order
func (k *KrakenClient) PlaceOrder(req *OrderRequest) (*OrderResponse, error) {
	endpoint := "/0/private/AddOrder"
	params := url.Values{}
	params.Set("pair", req.Pair)
	params.Set("type", req.Type)
	params.Set("ordertype", req.OrderType)
	params.Set("volume", req.Volume.String())
	if !req.Price.IsZero() {
		params.Set("price", req.Price.String())
	}

	resp, err := k.makeRequest("POST", endpoint, params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	var response OrderResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	if len(response.Error) > 0 {
		return nil, fmt.Errorf("kraken API error: %v", response.Error)
	}

	logrus.WithFields(logrus.Fields{
		"pair":   req.Pair,
		"type":   req.Type,
		"volume": req.Volume,
		"price":  req.Price,
	}).Info("Order placed successfully")

	return &response, nil
}

// makeRequest makes an HTTP request to Kraken API
func (k *KrakenClient) makeRequest(method, endpoint string, params url.Values, private bool) ([]byte, error) {
	var req *http.Request
	var err error

	fullURL := k.baseURL + endpoint

	if method == "GET" {
		if len(params) > 0 {
			fullURL += "?" + params.Encode()
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		req, err = http.NewRequest(method, fullURL, bytes.NewBufferString(params.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if private {
		nonce := strconv.FormatInt(time.Now().UnixNano(), 10)
		params.Set("nonce", nonce)

		signature, err := k.generateSignature(endpoint, params.Encode(), nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signature: %w", err)
		}

		req.Header.Set("API-Key", k.apiKey)
		req.Header.Set("API-Sign", signature)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// generateSignature generates the required signature for private API calls
func (k *KrakenClient) generateSignature(endpoint, postData, nonce string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(k.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	hash := sha256.Sum256([]byte(nonce + postData))
	hmacHash := hmac.New(sha512.New, decoded)
	hmacHash.Write([]byte(endpoint))
	hmacHash.Write(hash[:])

	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil)), nil
}
