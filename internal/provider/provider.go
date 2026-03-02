package provider

import (
	"context"
	"net/http"
)

type PriceProvider interface {
	Name() string
	FetchPrice(ctx context.Context) (float64, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
