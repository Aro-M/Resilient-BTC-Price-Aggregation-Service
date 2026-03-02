package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type krakenResponse struct {
	Error  []string `json:"error"`
	Result map[string]struct {
		C []string `json:"c"`
	} `json:"result"`
}

type krakenProvider struct {
	client HTTPClient
	apiURL string
}

func NewKraken(client HTTPClient, apiURL string) PriceProvider {
	return &krakenProvider{
		client: client,
		apiURL: apiURL,
	}
}
func (p *krakenProvider) Name() string { return "kraken" }

func (p *krakenProvider) FetchPrice(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("kraken: build request: %w", err)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("kraken: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("kraken: unexpected status %d", resp.StatusCode)
	}
	var result krakenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("kraken: decode: %w", err)
	}
	if len(result.Error) > 0 {
		return 0, fmt.Errorf("kraken: API error: %v", result.Error)
	}
	for _, ticker := range result.Result {
		if len(ticker.C) == 0 {
			return 0, fmt.Errorf("kraken: empty close price array")
		}
		price, err := strconv.ParseFloat(ticker.C[0], 64)
		if err != nil {
			return 0, fmt.Errorf("kraken: parse price %q: %w", ticker.C[0], err)
		}
		if price <= 0 {
			return 0, fmt.Errorf("kraken: non-positive price %.2f", price)
		}
		return price, nil
	}
	return 0, fmt.Errorf("kraken: no ticker data in response")
}
