package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// coinDeskResponse matches the custom data-api.coindesk.com format
type coinDeskResponse struct {
	Data map[string]struct {
		Price float64 `json:"PRICE"`
	} `json:"Data"`
}

type coinDeskProvider struct {
	client HTTPClient
	apiURL string
}

func NewCoinDesk(client HTTPClient, apiURL string) PriceProvider {
	return &coinDeskProvider{
		client: client,
		apiURL: apiURL,
	}
}
func (p *coinDeskProvider) Name() string { return "coindesk" }

func (p *coinDeskProvider) FetchPrice(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("coindesk: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("coindesk: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coindesk: unexpected status %d", resp.StatusCode)
	}
	var result coinDeskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("coindesk: decode: %w", err)
	}
	var price float64
	// Because the pair name (e.g., BTC-USD) is dynamic in the map key,
	// we will just grab the first valid price we find from the Data object.
	for _, pairData := range result.Data {
		if pairData.Price > 0 {
			price = pairData.Price
			break
		}
	}

	if price <= 0 {
		return 0, fmt.Errorf("coindesk: non-positive or missing price in response")
	}
	return price, nil
}
