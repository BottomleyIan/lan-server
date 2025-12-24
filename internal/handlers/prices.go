package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

const metalsPriceBaseURL = "https://api.gold-api.com/price/"
const metalsPriceCacheTTL = time.Minute

const currencyAPIBase = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies/usd.min.json"
const currencyAPIFallback = "https://latest.currency-api.pages.dev/v1/currencies/usd.min.json"

var metalsPricesCache = struct {
	mu        sync.Mutex
	expiresAt time.Time
	data      MetalsPricesDTO
}{}

// GetMetalPrices godoc
// @Summary Get current gold and silver prices
// @Tags prices
// @Produce json
// @Success 200 {object} MetalsPricesDTO
// @Router /prices/metals [get]
func (h *Handlers) GetMetalPrices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if cached, ok := getMetalsPricesCache(time.Now()); ok {
		writeJSON(w, cached)
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}

	gold, err := fetchMetalPrice(r.Context(), client, "XAU")
	if err != nil {
		http.Error(w, "price lookup failed", http.StatusBadGateway)
		return
	}

	silver, err := fetchMetalPrice(r.Context(), client, "XAG")
	if err != nil {
		http.Error(w, "price lookup failed", http.StatusBadGateway)
		return
	}

	usdToGBP, err := fetchUSDtoGBP(r.Context(), client)
	if err != nil {
		http.Error(w, "price lookup failed", http.StatusBadGateway)
		return
	}

	gold.GBP = gold.USD * usdToGBP
	silver.GBP = silver.USD * usdToGBP

	response := MetalsPricesDTO{
		Gold:   gold,
		Silver: silver,
	}

	setMetalsPricesCache(time.Now(), response)
	writeJSON(w, response)
}

type metalPriceAPIResponse struct {
	Name              string  `json:"name"`
	Price             float64 `json:"price"`
	Symbol            string  `json:"symbol"`
	UpdatedAt         string  `json:"updatedAt"`
	UpdatedAtReadable string  `json:"updatedAtReadable"`
}

func fetchMetalPrice(ctx context.Context, client *http.Client, symbol string) (MetalPriceDTO, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metalsPriceBaseURL+symbol, nil)
	if err != nil {
		return MetalPriceDTO{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return MetalPriceDTO{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MetalPriceDTO{}, errors.New("unexpected status")
	}

	var payload metalPriceAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return MetalPriceDTO{}, err
	}

	return MetalPriceDTO{
		Name:              payload.Name,
		USD:               payload.Price,
		GBP:               0,
		Symbol:            payload.Symbol,
		UpdatedAt:         payload.UpdatedAt,
		UpdatedAtReadable: payload.UpdatedAtReadable,
	}, nil
}

type currencyAPIResponse struct {
	USD map[string]float64 `json:"usd"`
}

func fetchUSDtoGBP(ctx context.Context, client *http.Client) (float64, error) {
	if rate, err := fetchUSDtoGBPFromURL(ctx, client, currencyAPIBase); err == nil {
		return rate, nil
	}
	return fetchUSDtoGBPFromURL(ctx, client, currencyAPIFallback)
}

func fetchUSDtoGBPFromURL(ctx context.Context, client *http.Client, url string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("unexpected status")
	}

	var payload currencyAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	rate, ok := payload.USD["gbp"]
	if !ok || rate == 0 {
		return 0, errors.New("missing gbp rate")
	}
	return rate, nil
}

func getMetalsPricesCache(now time.Time) (MetalsPricesDTO, bool) {
	metalsPricesCache.mu.Lock()
	defer metalsPricesCache.mu.Unlock()
	if now.Before(metalsPricesCache.expiresAt) {
		return metalsPricesCache.data, true
	}
	return MetalsPricesDTO{}, false
}

func setMetalsPricesCache(now time.Time, data MetalsPricesDTO) {
	metalsPricesCache.mu.Lock()
	defer metalsPricesCache.mu.Unlock()
	metalsPricesCache.data = data
	metalsPricesCache.expiresAt = now.Add(metalsPriceCacheTTL)
}
