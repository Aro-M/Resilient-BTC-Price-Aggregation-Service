package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type coinbaseResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

type coinbaseProvider struct {
	client HTTPClient
	apiURL string
}

func NewCoinbase(client HTTPClient, apiURL string) PriceProvider {
	return &coinbaseProvider{
		client: client,
		apiURL: apiURL,
	}
}
func (p *coinbaseProvider) Name() string { return "coinbase" }

func (p *coinbaseProvider) FetchPrice(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("coinbase: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("coinbase: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coinbase: unexpected status %d", resp.StatusCode)
	}
	var result coinbaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("coinbase: decode: %w", err)
	}
	price, err := strconv.ParseFloat(result.Data.Amount, 64)
	if err != nil {
		return 0, fmt.Errorf("coinbase: parse amount %q: %w", result.Data.Amount, err)
	}
	if price <= 0 {
		return 0, fmt.Errorf("coinbase: non-positive price %.2f", price)
	}
	return price, nil
}
